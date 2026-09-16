package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock Patient Repository
type mockPatientRepo struct {
	patients      []domain.Patient
	searchResults []domain.Patient
	searchErr     error
	createErr     error
}

func newMockPatientRepo() *mockPatientRepo {
	return &mockPatientRepo{
		patients: make([]domain.Patient, 0),
	}
}

func (m *mockPatientRepo) Search(hospitalID uuid.UUID, query *domain.PatientSearchQuery) ([]domain.Patient, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.searchResults != nil {
		return m.searchResults, nil
	}
	results := make([]domain.Patient, 0)
	for _, p := range m.patients {
		// Strict hospital_id data isolation
		if p.HospitalID != hospitalID {
			continue
		}

		if query != nil {
			if query.NationalID != "" && p.NationalID != query.NationalID {
				continue
			}
			if query.PassportID != "" && p.PassportID != query.PassportID {
				continue
			}
			if query.FirstName != "" && !strings.Contains(strings.ToLower(p.FirstNameTH), strings.ToLower(query.FirstName)) && !strings.Contains(strings.ToLower(p.FirstNameEN), strings.ToLower(query.FirstName)) {
				continue
			}
			if query.LastName != "" && !strings.Contains(strings.ToLower(p.LastNameTH), strings.ToLower(query.LastName)) && !strings.Contains(strings.ToLower(p.LastNameEN), strings.ToLower(query.LastName)) {
				continue
			}
			if query.PhoneNumber != "" && p.PhoneNumber != query.PhoneNumber {
				continue
			}
		}

		results = append(results, p)
	}
	return results, nil
}

func (m *mockPatientRepo) Create(p *domain.Patient) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.patients = append(m.patients, *p)
	return nil
}

func (m *mockPatientRepo) FindByNationalID(hospitalID uuid.UUID, nationalID string) (*domain.Patient, error) {
	for _, p := range m.patients {
		if p.HospitalID == hospitalID && p.NationalID == nationalID {
			return &p, nil
		}
	}
	return nil, nil
}

func (m *mockPatientRepo) FindByPassportID(hospitalID uuid.UUID, passportID string) (*domain.Patient, error) {
	for _, p := range m.patients {
		if p.HospitalID == hospitalID && p.PassportID == passportID {
			return &p, nil
		}
	}
	return nil, nil
}

func (m *mockPatientRepo) FindByHN(hospitalID uuid.UUID, hn string) (*domain.Patient, error) {
	for _, p := range m.patients {
		if p.HospitalID == hospitalID && p.PatientHN == hn {
			return &p, nil
		}
	}
	return nil, nil
}

// Mock HIS Client
type mockHISClient struct {
	patients  map[string]*domain.HospitalAPatientResponse
	searchErr error
}

func (m *mockHISClient) SearchPatient(id string) (*domain.HospitalAPatientResponse, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if p, ok := m.patients[id]; ok {
		return p, nil
	}
	return nil, nil
}

func TestPatientUseCase_DataIsolationAndHIS(t *testing.T) {
	repo := newMockPatientRepo()
	hisClient := &mockHISClient{
		patients: make(map[string]*domain.HospitalAPatientResponse),
	}

	uc := NewPatientUseCase(repo, hisClient)

	hospitalA_ID := uuid.New()
	hospitalB_ID := uuid.New()

	// Seed patient 1 for Hospital A
	patientA := domain.Patient{
		ID:          uuid.New(),
		HospitalID:  hospitalA_ID,
		PatientHN:   "HN-A-001",
		NationalID:  "1100100111111",
		PassportID:  "PA123456",
		FirstNameTH: "สมชาย",
		LastNameTH:  "ใจดี",
		FirstNameEN: "Somchai",
		LastNameEN:  "Jaidee",
		PhoneNumber: "0812345678",
	}
	_ = repo.Create(&patientA)

	// Seed patient 2 for Hospital B
	patientB := domain.Patient{
		ID:          uuid.New(),
		HospitalID:  hospitalB_ID,
		PatientHN:   "HN-B-999",
		NationalID:  "2200200222222",
		FirstNameTH: "สมศรี",
		LastNameTH:  "มีสุข",
		FirstNameEN: "Somsri",
		LastNameEN:  "Meesuk",
		PhoneNumber: "0898765432",
	}
	_ = repo.Create(&patientB)

	// 1. Positive Test: Hospital A searches and finds patient A
	resA, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		NationalID: "1100100111111",
	})
	assert.NoError(t, err)
	assert.Len(t, resA, 1)
	assert.Equal(t, "HN-A-001", resA[0].PatientHN)
	assert.Equal(t, hospitalA_ID, resA[0].HospitalID)

	// 2. Data Isolation Test (CRITICAL SECURITY): Hospital A searches for patient B's National ID
	// Must return EMPTY results, preserving privacy and multi-tenancy
	resCrossIsolation, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		NationalID: "2200200222222",
	})
	assert.NoError(t, err)
	assert.Empty(t, resCrossIsolation, "Hospital A must not see patients belonging to Hospital B")

	// 3. External HIS Synchronization Test
	// Add mock patient in Hospital A HIS
	hisClient.patients["3300300333333"] = &domain.HospitalAPatientResponse{
		PatientHN:   "HN-EXT-777",
		NationalID:  "3300300333333",
		FirstNameTH: "ประสิทธิ์",
		LastNameTH:  "เจริญสุข",
		FirstNameEN: "Prasit",
		LastNameEN:  "Charoensuk",
	}

	resHIS, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		NationalID: "3300300333333",
	})
	assert.NoError(t, err)
	assert.Len(t, resHIS, 1)
	assert.Equal(t, "HN-EXT-777", resHIS[0].PatientHN)
	assert.Equal(t, hospitalA_ID, resHIS[0].HospitalID)

	// 4. Passport Data Isolation Test: Hospital B patient with passport
	patientB_Passport := domain.Patient{
		ID:          uuid.New(),
		HospitalID:  hospitalB_ID,
		PatientHN:   "HN-B-888",
		PassportID:  "AB9876543",
		FirstNameEN: "John",
		LastNameEN:  "Doe",
	}
	_ = repo.Create(&patientB_Passport)

	// Hospital A searches for Hospital B's passport -> MUST BE EMPTY
	resPassportCross, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		PassportID: "AB9876543",
	})
	assert.NoError(t, err)
	assert.Empty(t, resPassportCross, "Hospital A must not see patients with Passport belonging to Hospital B")

	// 5. Multi-field Search Tests (First Name, Last Name, Phone)
	resByNameTH, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		FirstName: "สมชาย",
	})
	assert.NoError(t, err)
	assert.True(t, len(resByNameTH) >= 1)

	resByNameEN, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		FirstName: "Somchai",
	})
	assert.NoError(t, err)
	assert.True(t, len(resByNameEN) >= 1)

	resByPhone, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		PhoneNumber: "0812345678",
	})
	assert.NoError(t, err)
	assert.Len(t, resByPhone, 1)

	// 6. External HIS with Passport ID
	hisClient.patients["P555666777"] = &domain.HospitalAPatientResponse{
		PatientHN:   "HN-EXT-PASSPORT",
		PassportID:  "P555666777",
		FirstNameEN: "Michael",
		LastNameEN:  "Scott",
	}
	resHISPassport, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		PassportID: "P555666777",
	})
	assert.NoError(t, err)
	assert.Len(t, resHISPassport, 1)
	assert.Equal(t, "HN-EXT-PASSPORT", resHISPassport[0].PatientHN)

	// 7. Patient already in local DB should not be duplicated when re-searched
	resAlreadyInDB, err := uc.SearchPatients(hospitalA_ID, &domain.PatientSearchQuery{
		NationalID: "1100100111111",
	})
	assert.NoError(t, err)
	assert.Len(t, resAlreadyInDB, 1)
}

func TestPatientUseCase_EdgeCasesAndErrors(t *testing.T) {
	hospitalID := uuid.New()

	// 1. Repo Search Error
	errRepo := newMockPatientRepo()
	errRepo.searchErr = errors.New("database connection failed")
	uc1 := NewPatientUseCase(errRepo, nil)
	_, err := uc1.SearchPatients(hospitalID, &domain.PatientSearchQuery{NationalID: "123"})
	assert.Error(t, err)
	assert.Equal(t, "database connection failed", err.Error())

	// 2. HIS Search Error (should log warning and continue without breaking)
	repo := newMockPatientRepo()
	errHISClient := &mockHISClient{searchErr: errors.New("upstream service unavailable")}
	uc2 := NewPatientUseCase(repo, errHISClient)
	res, err := uc2.SearchPatients(hospitalID, &domain.PatientSearchQuery{NationalID: "999"})
	assert.NoError(t, err)
	assert.Empty(t, res)

	// 3. HIS Patient already in list by NationalID
	repoWithNational := newMockPatientRepo()
	repoWithNational.searchResults = []domain.Patient{
		{
			ID:         uuid.New(),
			HospitalID: hospitalID,
			NationalID: "NAT_DUP",
		},
	}
	dupHisClientNational := &mockHISClient{
		patients: map[string]*domain.HospitalAPatientResponse{
			"SEARCH_NAT": {
				NationalID: "NAT_DUP",
			},
		},
	}
	uc3 := NewPatientUseCase(repoWithNational, dupHisClientNational)
	res3, err3 := uc3.SearchPatients(hospitalID, &domain.PatientSearchQuery{PassportID: "SEARCH_NAT"})
	assert.NoError(t, err3)
	assert.Len(t, res3, 1)

	// 4. HIS Patient already in list by PassportID
	repoWithPassport := newMockPatientRepo()
	repoWithPassport.searchResults = []domain.Patient{
		{
			ID:         uuid.New(),
			HospitalID: hospitalID,
			PassportID: "PASS_DUP",
		},
	}
	dupHisClientPassport := &mockHISClient{
		patients: map[string]*domain.HospitalAPatientResponse{
			"SEARCH_PASS": {
				PassportID: "PASS_DUP",
			},
		},
	}
	uc4 := NewPatientUseCase(repoWithPassport, dupHisClientPassport)
	res4, err4 := uc4.SearchPatients(hospitalID, &domain.PatientSearchQuery{NationalID: "SEARCH_PASS"})
	assert.NoError(t, err4)
	assert.Len(t, res4, 1)

	// 5. Repo Create Error when persisting HIS Patient (saveErr != nil fallback branch)
	repoFailCreate := newMockPatientRepo()
	repoFailCreate.createErr = errors.New("db disk full")
	hisClientNew := &mockHISClient{
		patients: map[string]*domain.HospitalAPatientResponse{
			"NEW_ID": {
				PatientHN:  "HN-NEW",
				NationalID: "NEW_ID",
			},
		},
	}
	uc5 := NewPatientUseCase(repoFailCreate, hisClientNew)
	res5, err5 := uc5.SearchPatients(hospitalID, &domain.PatientSearchQuery{NationalID: "NEW_ID"})
	assert.NoError(t, err5)
	assert.Len(t, res5, 1)
	assert.Equal(t, "HN-NEW", res5[0].PatientHN)
}
