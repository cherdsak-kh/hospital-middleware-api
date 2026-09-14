package middleware

import (
	"net/http"
	"strings"

	"github.com/cherdsak-kh/hospital-middleware-api/pkg/response"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT Bearer tokens and injects claims into Gin context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.JSONError(c, http.StatusUnauthorized, "Authorization header required", "Missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.JSONError(c, http.StatusUnauthorized, "Invalid authorization format", "Authorization format must be Bearer <token>")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			response.JSONError(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Inject verified claims into context
		c.Set("staff_id", claims.StaffID)
		c.Set("username", claims.Username)
		c.Set("hospital_id", claims.HospitalID)

		c.Next()
	}
}
