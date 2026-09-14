package usecase

import (
	"strings"
	"testing"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock Hospital Repository
type mockHospitalRepo struct {
	hospitals map[string]*domain.Hospital
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
	staffs []*domain.Staff
}

func newMockStaffRepo() *mockStaffRepo {
	return &mockStaffRepo{
		staffs: make([]*domain.Staff, 0),
	}
}

func (m *mockStaffRepo) Create(s *domain.Staff) error {
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
	for _, s := range m.staffs {
		if strings.EqualFold(s.Username, username) {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockStaffRepo) FindByUsernameAndHospital(username string, hospitalID uuid.UUID) (*domain.Staff, error) {
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
		Hospital: "Hospital B", // Staff Alice is in Hospital A, not Hospital B
	}
	_, wrongHospErr := uc.LoginStaff(wrongHospReq)
	assert.Error(t, wrongHospErr)
}
