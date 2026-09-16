package http

import (
	"bytes"
	"encoding/json"
	"errors"
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
type mockStaffUseCase struct{}

func (m *mockStaffUseCase) CreateStaff(req *domain.StaffCreateRequest) (*domain.Staff, error) {
	if req.Username == "duplicate_user" {
		return nil, usecase.ErrUsernameTaken
	}
	if req.Username == "server_error" {
		return nil, errors.New("database connection failure")
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
	if req.Username == "server_error" {
		return nil, errors.New("internal server error in staff login")
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
	if query != nil && query.NationalID == "FAIL_USECASE" {
		return nil, errors.New("database query failed in usecase")
	}
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

func TestRouter_OptionsAndProductionMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prodCfg := &config.Config{
		JWTSecret: "secret",
		AppEnv:    "production",
	}
	prodRouter := SetupRouter(prodCfg, NewStaffHandler(&mockStaffUseCase{}), NewPatientHandler(&mockPatientUseCase{}))
	assert.NotNil(t, prodRouter)

	// Test OPTIONS request (CORS Preflight)
	reqOptions, _ := http.NewRequest(http.MethodOptions, "/patient/search", nil)
	wOptions := httptest.NewRecorder()
	prodRouter.ServeHTTP(wOptions, reqOptions)
	assert.Equal(t, http.StatusNoContent, wOptions.Code)
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

	// 4. Negative: Internal server error (500)
	errPayload := domain.StaffCreateRequest{
		Username: "server_error",
		Password: "Password123",
		Hospital: "Hospital A",
	}
	errBody, _ := json.Marshal(errPayload)
	reqErr, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(errBody))
	reqErr.Header.Set("Content-Type", "application/json")
	wErr := httptest.NewRecorder()
	router.ServeHTTP(wErr, reqErr)
	assert.Equal(t, http.StatusInternalServerError, wErr.Code)
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

	// 3. Negative: Hospital not found (401)
	wrongHospPayload := domain.StaffLoginRequest{
		Username: "valid_user",
		Password: "Password123",
		Hospital: "wrong_hospital",
	}
	wrongHospBody, _ := json.Marshal(wrongHospPayload)
	reqHosp, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(wrongHospBody))
	reqHosp.Header.Set("Content-Type", "application/json")
	wHosp := httptest.NewRecorder()
	router.ServeHTTP(wHosp, reqHosp)
	assert.Equal(t, http.StatusUnauthorized, wHosp.Code)

	// 4. Negative: Internal server error on login (500)
	errPayload := domain.StaffLoginRequest{
		Username: "server_error",
		Password: "Password123",
		Hospital: "Hospital A",
	}
	errBody, _ := json.Marshal(errPayload)
	reqErr, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(errBody))
	reqErr.Header.Set("Content-Type", "application/json")
	wErr := httptest.NewRecorder()
	router.ServeHTTP(wErr, reqErr)
	assert.Equal(t, http.StatusInternalServerError, wErr.Code)

	// 5. Negative: Malformed body (400)
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

	// 3. POST /patient/search with malformed JSON body (falls back to query parameters)
	reqPostBadJSON, _ := http.NewRequest(http.MethodPost, "/patient/search?national_id=1100100111111", bytes.NewBuffer([]byte("{broken}")))
	reqPostBadJSON.Header.Set("Authorization", "Bearer "+token)
	reqPostBadJSON.Header.Set("Content-Type", "application/json")
	wPostBadJSON := httptest.NewRecorder()
	router.ServeHTTP(wPostBadJSON, reqPostBadJSON)
	assert.Equal(t, http.StatusOK, wPostBadJSON.Code)

	// 4. Negative: Usecase error returns 500
	reqFail, _ := http.NewRequest(http.MethodGet, "/patient/search?national_id=FAIL_USECASE", nil)
	reqFail.Header.Set("Authorization", "Bearer "+token)
	wFail := httptest.NewRecorder()
	router.ServeHTTP(wFail, reqFail)
	assert.Equal(t, http.StatusInternalServerError, wFail.Code)

	// 5. Negative: Missing token (401 Unauthorized)
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/patient/search?national_id=1100100111111", nil)
	wNoAuth := httptest.NewRecorder()
	router.ServeHTTP(wNoAuth, reqNoAuth)
	assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)
}

func TestPatientHandler_ContextHospitalIDVariations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	patientHandler := NewPatientHandler(&mockPatientUseCase{})

	// 1. Missing hospital_id in context
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodGet, "/patient/search", nil)
	patientHandler.SearchPatients(c1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// 2. String valid hospital_id in context
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodGet, "/patient/search", nil)
	c2.Set("hospital_id", uuid.New().String())
	patientHandler.SearchPatients(c2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 3. String invalid UUID in context
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request, _ = http.NewRequest(http.MethodGet, "/patient/search", nil)
	c3.Set("hospital_id", "invalid-uuid-string")
	patientHandler.SearchPatients(c3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	// 4. Unknown type (int) in context
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request, _ = http.NewRequest(http.MethodGet, "/patient/search", nil)
	c4.Set("hospital_id", 12345)
	patientHandler.SearchPatients(c4)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)
}
