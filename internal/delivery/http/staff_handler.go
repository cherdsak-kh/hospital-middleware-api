package http

import (
	"errors"
	"net/http"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/usecase"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// StaffHandler handles HTTP requests for staff management
type StaffHandler struct {
	staffUseCase domain.StaffUseCase
}

// NewStaffHandler creates a new staff handler instance
func NewStaffHandler(staffUseCase domain.StaffUseCase) *StaffHandler {
	return &StaffHandler{staffUseCase: staffUseCase}
}

// CreateStaff godoc
// @Summary      Create a new hospital staff member
// @Description  Registers a new staff member with encrypted login credentials and hospital affiliation
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        request body domain.StaffCreateRequest true "Staff creation payload"
// @Success      201 {object} response.StandardResponse "Staff account created successfully"
// @Failure      400 {object} response.StandardResponse "Invalid request payload"
// @Failure      409 {object} response.StandardResponse "Username is already taken"
// @Failure      500 {object} response.StandardResponse "Internal server error"
// @Router       /staff/create [post]
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req domain.StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "Validation error: invalid request payload", err.Error())
		return
	}

	staff, err := h.staffUseCase.CreateStaff(&req)
	if err != nil {
		if errors.Is(err, usecase.ErrUsernameTaken) {
			response.JSONError(c, http.StatusConflict, "Username is already taken", err.Error())
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "Failed to create staff account", err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusCreated, "Staff account created successfully", gin.H{
		"id":          staff.ID,
		"username":    staff.Username,
		"hospital_id": staff.HospitalID,
		"full_name":   staff.FullName,
		"created_at":  staff.CreatedAt,
	})
}

// LoginStaff godoc
// @Summary      Hospital staff login
// @Description  Authenticates hospital staff and returns a signed JWT token containing hospital context
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        request body domain.StaffLoginRequest true "Staff login credentials"
// @Success      200 {object} response.StandardResponse{data=domain.StaffLoginResponse} "Authentication successful"
// @Failure      400 {object} response.StandardResponse "Invalid login payload"
// @Failure      401 {object} response.StandardResponse "Authentication failed"
// @Failure      500 {object} response.StandardResponse "Internal server error"
// @Router       /staff/login [post]
func (h *StaffHandler) LoginStaff(c *gin.Context) {
	var req domain.StaffLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "Validation error: invalid login payload", err.Error())
		return
	}

	loginResp, err := h.staffUseCase.LoginStaff(&req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAuth) || errors.Is(err, usecase.ErrHospitalNotFound) {
			response.JSONError(c, http.StatusUnauthorized, "Authentication failed", "Invalid username, password, or hospital")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "Authentication error", err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Authentication successful", loginResp)
}
