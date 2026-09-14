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

// SearchPatients handles GET and POST /patient/search
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
