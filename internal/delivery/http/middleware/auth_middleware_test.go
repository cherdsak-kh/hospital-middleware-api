package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_auth_secret_xyz"

	setupRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(AuthMiddleware(secret))
		r.GET("/protected", func(c *gin.Context) {
			hospitalID, _ := c.Get("hospital_id")
			staffID, _ := c.Get("staff_id")
			c.JSON(http.StatusOK, gin.H{
				"hospital_id": hospitalID,
				"staff_id":    staffID,
			})
		})
		return r
	}

	r := setupRouter()

	staffID := uuid.New()
	hospitalID := uuid.New()
	validToken, err := utils.GenerateToken(staffID, hospitalID, "staff_user", secret, 2)
	assert.NoError(t, err)

	// Positive test: Valid Bearer token
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Negative test 1: Missing Authorization header
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	wNoAuth := httptest.NewRecorder()
	r.ServeHTTP(wNoAuth, reqNoAuth)
	assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)

	// Negative test 2: Invalid scheme (Basic instead of Bearer)
	reqBasic, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	reqBasic.Header.Set("Authorization", "Basic "+validToken)
	wBasic := httptest.NewRecorder()
	r.ServeHTTP(wBasic, reqBasic)
	assert.Equal(t, http.StatusUnauthorized, wBasic.Code)

	// Negative test 3: Tampered or invalid token
	reqInvalidToken, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	reqInvalidToken.Header.Set("Authorization", "Bearer invalid.jwt.string")
	wInvalidToken := httptest.NewRecorder()
	r.ServeHTTP(wInvalidToken, reqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, wInvalidToken.Code)
}
