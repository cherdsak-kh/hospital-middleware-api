package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cherdsak-kh/hospital-middleware-api/config"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/usecase"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Mock Staff UseCase
type mockStaffUseCase struct {
	staffs map[string]*domain.Staff
}

func (m *mockStaffUseCase) CreateStaff(req *domain.StaffCreateRequest) (*domain.Staff, error) {
	if req.Username == "duplicate_user" {
		return nil, usecase.ErrUsernameTaken
	}
	s := &domain.Staff{
		ID:         uuid.New(),
		HospitalID: uuid.New(),
		Username:   req.Username,
		FullName:   req.FullName,
	}
	return s, nil
}

func (m *mockStaffUseCase) LoginStaff(req *domain.StaffLoginRequest) (*domain.StaffLoginResponse, error) {
	if req.Username == "wrong_user" || req.Password == "wrong_password" {
		return nil, usecase.ErrInvalidAuth
	}
	if req.Hospital == "wrong_hospital" {
		return nil, usecase.ErrHospitalNotFound
	}
	return &domain.StaffLoginResponse{
		Token:        "mock-jwt-token-12345",
		StaffID:      uuid.New(),
		Username:     req.Username,
		HospitalID:   uuid.New(),
		HospitalName: req.Hospital,
	}, nil
}

// Mock Patient UseCase
type mockPatientUseCase struct{}

func (m *mockPatientUseCase) SearchPatients(hospitalID uuid.UUID, query *domain.PatientSearchQuery) ([]domain.PatientResponse, error) {
	return []domain.PatientResponse{
		{
			ID:          uuid.New(),
			HospitalID:  hospitalID,
			PatientHN:   "HN-TEST-001",
			NationalID:  query.NationalID,
			FirstNameTH: "ทดสอบ",
			LastNameTH:  "ระบบ",
		},
	}, nil
}

func setupTestRouter() (*gin.Engine, string, uuid.UUID, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test_secret_for_handlers"
	cfg := &config.Config{
		JWTSecret: jwtSecret,
		AppEnv:    "test",
	}

	staffHandler := NewStaffHandler(&mockStaffUseCase{})
	patientHandler := NewPatientHandler(&mockPatientUseCase{})

	router := SetupRouter(cfg, staffHandler, patientHandler)
	staffID := uuid.New()
	hospitalID := uuid.New()

	return router, jwtSecret, staffID, hospitalID
}

func TestRootAndHealthEndpoints(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// GET /
	reqRoot, _ := http.NewRequest(http.MethodGet, "/", nil)
	wRoot := httptest.NewRecorder()
	router.ServeHTTP(wRoot, reqRoot)
	assert.Equal(t, http.StatusOK, wRoot.Code)
	assert.Contains(t, wRoot.Body.String(), "hospital-middleware-api")

	// GET /health
	reqHealth, _ := http.NewRequest(http.MethodGet, "/health", nil)
	wHealth := httptest.NewRecorder()
	router.ServeHTTP(wHealth, reqHealth)
	assert.Equal(t, http.StatusOK, wHealth.Code)
	assert.Contains(t, wHealth.Body.String(), "ok")
	assert.Contains(t, wHealth.Body.String(), "uptime")
}

func TestStaffHandler_CreateStaff(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// 1. Positive: Successfully create staff (201)
	payload := domain.StaffCreateRequest{
		Username: "new_doctor",
		Password: "Password123",
		Hospital: "Hospital A",
		FullName: "Dr. New",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "new_doctor")

	// 2. Negative: Invalid JSON payload (400)
	reqBad, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer([]byte("{invalid-json}")))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	router.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusBadRequest, wBad.Code)

	// 3. Negative: Duplicate username (409 Conflict)
	dupPayload := domain.StaffCreateRequest{
		Username: "duplicate_user",
		Password: "Password123",
		Hospital: "Hospital A",
	}
	dupBody, _ := json.Marshal(dupPayload)
	reqDup, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(dupBody))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	router.ServeHTTP(wDup, reqDup)
	assert.Equal(t, http.StatusConflict, wDup.Code)
}

func TestStaffHandler_LoginStaff(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// 1. Positive: Successful login (200)
	loginPayload := domain.StaffLoginRequest{
		Username: "valid_doctor",
		Password: "Password123",
		Hospital: "Hospital A",
	}
	body, _ := json.Marshal(loginPayload)
	req, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "mock-jwt-token-12345")

	// 2. Negative: Invalid credentials (401)
	wrongPayload := domain.StaffLoginRequest{
		Username: "wrong_user",
		Password: "wrong_password",
		Hospital: "Hospital A",
	}
	wrongBody, _ := json.Marshal(wrongPayload)
	reqWrong, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(wrongBody))
	reqWrong.Header.Set("Content-Type", "application/json")
	wWrong := httptest.NewRecorder()
	router.ServeHTTP(wWrong, reqWrong)
	assert.Equal(t, http.StatusUnauthorized, wWrong.Code)

	// 3. Negative: Malformed body (400)
	reqBad, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer([]byte("{bad-json}")))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	router.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusBadRequest, wBad.Code)
}

func TestPatientHandler_SearchPatients(t *testing.T) {
	router, secret, staffID, hospitalID := setupTestRouter()

	token, err := utils.GenerateToken(staffID, hospitalID, "test_doctor", secret, 2)
	assert.NoError(t, err)

	// 1. Positive: GET /patient/search with valid token
	reqGet, _ := http.NewRequest(http.MethodGet, "/patient/search?national_id=1100100111111", nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	assert.Contains(t, wGet.Body.String(), "HN-TEST-001")

	// 2. Positive: POST /patient/search with valid token
	postQuery := domain.PatientSearchQuery{
		NationalID: "1100100111111",
		FirstName:  "ทดสอบ",
	}
	postBody, _ := json.Marshal(postQuery)
	reqPost, _ := http.NewRequest(http.MethodPost, "/patient/search", bytes.NewBuffer(postBody))
	reqPost.Header.Set("Authorization", "Bearer "+token)
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	router.ServeHTTP(wPost, reqPost)
	assert.Equal(t, http.StatusOK, wPost.Code)
	assert.Contains(t, wPost.Body.String(), "HN-TEST-001")

	// 3. Negative: Missing token (401 Unauthorized)
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/patient/search?national_id=1100100111111", nil)
	wNoAuth := httptest.NewRecorder()
	router.ServeHTTP(wNoAuth, reqNoAuth)
	assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)
}
