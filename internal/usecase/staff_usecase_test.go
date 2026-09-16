package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock Hospital Repository
type mockHospitalRepo struct {
	hospitals         map[string]*domain.Hospital
	findOrCreateErr   error
	findByCodeNameErr error
}

func newMockHospitalRepo() *mockHospitalRepo {
	return &mockHospitalRepo{
		hospitals: make(map[string]*domain.Hospital),
	}
}

func (m *mockHospitalRepo) FindByID(id uuid.UUID) (*domain.Hospital, error) {
	for _, h := range m.hospitals {
		if h.ID == id {
			return h, nil
		}
	}
	return nil, nil
}

func (m *mockHospitalRepo) FindByCode(code string) (*domain.Hospital, error) {
	if h, ok := m.hospitals[strings.ToLower(code)]; ok {
		return h, nil
	}
	return nil, nil
}

func (m *mockHospitalRepo) FindByName(name string) (*domain.Hospital, error) {
	for _, h := range m.hospitals {
		if strings.EqualFold(h.Name, name) {
			return h, nil
		}
	}
	return nil, nil
}

func (m *mockHospitalRepo) FindByCodeOrName(identifier string) (*domain.Hospital, error) {
	if m.findByCodeNameErr != nil {
		return nil, m.findByCodeNameErr
	}
	lower := strings.ToLower(strings.TrimSpace(identifier))
	for _, h := range m.hospitals {
		if strings.ToLower(h.Code) == lower || strings.ToLower(h.Name) == lower {
			return h, nil
		}
	}
	return nil, nil
}

func (m *mockHospitalRepo) Create(h *domain.Hospital) error {
	m.hospitals[strings.ToLower(h.Code)] = h
	return nil
}

func (m *mockHospitalRepo) FindOrCreate(identifier string) (*domain.Hospital, error) {
	if m.findOrCreateErr != nil {
		return nil, m.findOrCreateErr
	}
	if h, _ := m.FindByCodeOrName(identifier); h != nil {
		return h, nil
	}
	cleanCode := strings.ToUpper(strings.ReplaceAll(identifier, " ", "_"))
	newHospital := &domain.Hospital{
		ID:   uuid.New(),
		Code: cleanCode,
		Name: identifier,
	}
	_ = m.Create(newHospital)
	return newHospital, nil
}

// Mock Staff Repository
type mockStaffRepo struct {
	staffs             []*domain.Staff
	findByUsernameErr  error
	findByStaffHospErr error
	createErr          error
}

func newMockStaffRepo() *mockStaffRepo {
	return &mockStaffRepo{
		staffs: make([]*domain.Staff, 0),
	}
}

func (m *mockStaffRepo) Create(s *domain.Staff) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.staffs = append(m.staffs, s)
	return nil
}

func (m *mockStaffRepo) FindByID(id uuid.UUID) (*domain.Staff, error) {
	for _, s := range m.staffs {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockStaffRepo) FindByUsername(username string) (*domain.Staff, error) {
	if m.findByUsernameErr != nil {
		return nil, m.findByUsernameErr
	}
	for _, s := range m.staffs {
		if strings.EqualFold(s.Username, username) {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockStaffRepo) FindByUsernameAndHospital(username string, hospitalID uuid.UUID) (*domain.Staff, error) {
	if m.findByStaffHospErr != nil {
		return nil, m.findByStaffHospErr
	}
	for _, s := range m.staffs {
		if strings.EqualFold(s.Username, username) && s.HospitalID == hospitalID {
			return s, nil
		}
	}
	return nil, nil
}

func TestStaffUseCase_CreateAndLogin(t *testing.T) {
	hospitalRepo := newMockHospitalRepo()
	staffRepo := newMockStaffRepo()
	jwtSecret := "test_secret_key"
	uc := NewStaffUseCase(staffRepo, hospitalRepo, jwtSecret, 24)

	// 1. Positive Test: Create staff successfully
	createReq := &domain.StaffCreateRequest{
		Username: "staff_alice",
		Password: "SecurePassword123",
		Hospital: "Hospital A",
		FullName: "Alice Smith",
	}
	staff, err := uc.CreateStaff(createReq)
	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "staff_alice", staff.Username)
	assert.True(t, utils.CheckPassword(staff.PasswordHash, "SecurePassword123"))

	// 2. Negative Test: Duplicate username creation
	duplicateStaff, dupErr := uc.CreateStaff(createReq)
	assert.Error(t, dupErr)
	assert.Nil(t, duplicateStaff)
	assert.Equal(t, ErrUsernameTaken, dupErr)

	// 3. Positive Test: Successful login
	loginReq := &domain.StaffLoginRequest{
		Username: "staff_alice",
		Password: "SecurePassword123",
		Hospital: "Hospital A",
	}
	loginResp, err := uc.LoginStaff(loginReq)
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.Token)
	assert.Equal(t, staff.ID, loginResp.StaffID)

	// 4. Negative Test: Wrong password
	wrongPassReq := &domain.StaffLoginRequest{
		Username: "staff_alice",
		Password: "WrongPassword!",
		Hospital: "Hospital A",
	}
	_, wrongPassErr := uc.LoginStaff(wrongPassReq)
	assert.Error(t, wrongPassErr)
	assert.Equal(t, ErrInvalidAuth, wrongPassErr)

	// 5. Negative Test: Hospital not found / cross-hospital unauthorized login
	wrongHospReq := &domain.StaffLoginRequest{
		Username: "staff_alice",
		Password: "SecurePassword123",
		Hospital: "Hospital B",
	}
	_, wrongHospErr := uc.LoginStaff(wrongHospReq)
	assert.Error(t, wrongHospErr)

	// 6. Negative Test: Non-existent user login
	nonExistentReq := &domain.StaffLoginRequest{
		Username: "staff_ghost",
		Password: "SecurePassword123",
		Hospital: "Hospital A",
	}
	_, ghostErr := uc.LoginStaff(nonExistentReq)
	assert.Error(t, ghostErr)
	assert.Equal(t, ErrInvalidAuth, ghostErr)

	// 7. Negative Test: Empty username registration
	emptyUserReq := &domain.StaffCreateRequest{
		Username: "   ",
		Password: "SecurePassword123",
		Hospital: "Hospital A",
	}
	_, emptyUserErr := uc.CreateStaff(emptyUserReq)
	assert.Error(t, emptyUserErr)

	// 8. Positive Test: NewStaffUseCase with default expiry fallback
	defaultUC := NewStaffUseCase(staffRepo, hospitalRepo, jwtSecret, 0)
	assert.NotNil(t, defaultUC)
}

func TestStaffUseCase_ErrorBranches(t *testing.T) {
	jwtSecret := "test_secret_key"

	// 1. staffRepo.FindByUsername error
	sRepoErr := newMockStaffRepo()
	sRepoErr.findByUsernameErr = errors.New("db error on find username")
	uc1 := NewStaffUseCase(sRepoErr, newMockHospitalRepo(), jwtSecret, 24)
	_, err1 := uc1.CreateStaff(&domain.StaffCreateRequest{Username: "user1", Password: "pwd", Hospital: "Hosp"})
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "failed to check existing staff")

	// 2. hospitalRepo.FindOrCreate error
	hRepoErr := newMockHospitalRepo()
	hRepoErr.findOrCreateErr = errors.New("db error on hospital create")
	uc2 := NewStaffUseCase(newMockStaffRepo(), hRepoErr, jwtSecret, 24)
	_, err2 := uc2.CreateStaff(&domain.StaffCreateRequest{Username: "user2", Password: "pwd", Hospital: "Hosp"})
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "failed to resolve hospital")

	// 3. Password > 72 bytes error
	uc3 := NewStaffUseCase(newMockStaffRepo(), newMockHospitalRepo(), jwtSecret, 24)
	_, err3 := uc3.CreateStaff(&domain.StaffCreateRequest{Username: "user3", Password: string(make([]byte, 100)), Hospital: "Hosp"})
	assert.Error(t, err3)
	assert.Contains(t, err3.Error(), "failed to hash password")

	// 4. staffRepo.Create error
	sRepoCreateErr := newMockStaffRepo()
	sRepoCreateErr.createErr = errors.New("db insert failure")
	uc4 := NewStaffUseCase(sRepoCreateErr, newMockHospitalRepo(), jwtSecret, 24)
	_, err4 := uc4.CreateStaff(&domain.StaffCreateRequest{Username: "user4", Password: "pwd", Hospital: "Hosp"})
	assert.Error(t, err4)
	assert.Contains(t, err4.Error(), "failed to create staff record")

	// 5. hospitalRepo.FindByCodeOrName error in Login
	hRepoFindErr := newMockHospitalRepo()
	hRepoFindErr.findByCodeNameErr = errors.New("db query hospital failed")
	uc5 := NewStaffUseCase(newMockStaffRepo(), hRepoFindErr, jwtSecret, 24)
	_, err5 := uc5.LoginStaff(&domain.StaffLoginRequest{Username: "user5", Password: "pwd", Hospital: "Hosp"})
	assert.Error(t, err5)
	assert.Contains(t, err5.Error(), "failed to find hospital")

	// 6. staffRepo.FindByUsernameAndHospital error in Login
	sRepoStaffErr := newMockStaffRepo()
	sRepoStaffErr.findByStaffHospErr = errors.New("db query staff failed")
	hRepoValid := newMockHospitalRepo()
	_ = hRepoValid.Create(&domain.Hospital{ID: uuid.New(), Code: "HOSP", Name: "Hosp"})
	uc6 := NewStaffUseCase(sRepoStaffErr, hRepoValid, jwtSecret, 24)
	_, err6 := uc6.LoginStaff(&domain.StaffLoginRequest{Username: "user6", Password: "pwd", Hospital: "Hosp"})
	assert.Error(t, err6)
	assert.Contains(t, err6.Error(), "failed to find staff")
}
