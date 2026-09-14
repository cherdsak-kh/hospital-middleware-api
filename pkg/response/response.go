package response

import (
	"github.com/gin-gonic/gin"
)

// StandardResponse defines the standard API response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// JSONSuccess sends a standardized successful JSON response
func JSONSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// JSONError sends a standardized error JSON response
func JSONError(c *gin.Context, statusCode int, message string, errDetail interface{}) {
	var errVal interface{} = errDetail
	if err, ok := errDetail.(error); ok && err != nil {
		errVal = err.Error()
	}

	c.JSON(statusCode, StandardResponse{
		Success: false,
		Message: message,
		Error:   errVal,
	})
}
