package http

import (
	"net/http"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PatientHandler handles HTTP requests for patient queries
type PatientHandler struct {
	patientUseCase domain.PatientUseCase
}

// NewPatientHandler creates a new patient handler instance
func NewPatientHandler(patientUseCase domain.PatientUseCase) *PatientHandler {
	return &PatientHandler{patientUseCase: patientUseCase}
}

// SearchPatients godoc
// @Summary      Search patient records
// @Description  Searches patient records by optional criteria with hospital data isolation
// @Tags         Patient
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        national_id query string false "National ID"
// @Param        passport_id query string false "Passport ID"
// @Param        first_name query string false "First Name (matches TH or EN)"
// @Param        middle_name query string false "Middle Name (matches TH or EN)"
// @Param        last_name query string false "Last Name (matches TH or EN)"
// @Param        date_of_birth query string false "Date of Birth (YYYY-MM-DD)"
// @Param        phone_number query string false "Phone Number"
// @Param        email query string false "Email Address"
// @Success      200 {object} response.StandardResponse{data=[]domain.PatientResponse} "Patients matching criteria"
// @Failure      400 {object} response.StandardResponse "Invalid query parameters"
// @Failure      401 {object} response.StandardResponse "Unauthorized or missing token"
// @Failure      500 {object} response.StandardResponse "Internal server error"
// @Router       /patient/search [get]
// @Router       /patient/search [post]
func (h *PatientHandler) SearchPatients(c *gin.Context) {
	// Extract verified hospital_id injected by AuthMiddleware
	hospitalIDVal, exists := c.Get("hospital_id")
	if !exists {
		response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "Hospital context not found")
		return
	}

	var hospitalID uuid.UUID
	switch v := hospitalIDVal.(type) {
	case uuid.UUID:
		hospitalID = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "Invalid hospital ID format in token")
			return
		}
		hospitalID = parsed
	default:
		response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "Unknown hospital ID type in token")
		return
	}

	var query domain.PatientSearchQuery
	// Support both GET (Query parameters) and POST (JSON/Form body)
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBindJSON(&query); err != nil {
			// If JSON parsing fails, fallback to binding query parameters
			_ = c.ShouldBindQuery(&query)
		}
	} else {
		if err := c.ShouldBindQuery(&query); err != nil {
			response.JSONError(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
			return
		}
	}

	patients, err := h.patientUseCase.SearchPatients(hospitalID, &query)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to search patients", err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Patients retrieved successfully", gin.H{
		"total":    len(patients),
		"patients": patients,
	})
}
