package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJSONSuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test JSONSuccess
	wSuccess := httptest.NewRecorder()
	cSuccess, _ := gin.CreateTestContext(wSuccess)
	JSONSuccess(cSuccess, http.StatusOK, "operation successful", gin.H{"id": 123})

	assert.Equal(t, http.StatusOK, wSuccess.Code)
	var respSuccess StandardResponse
	err := json.Unmarshal(wSuccess.Body.Bytes(), &respSuccess)
	assert.NoError(t, err)
	assert.True(t, respSuccess.Success)
	assert.Equal(t, "operation successful", respSuccess.Message)

	// Test JSONError with error type
	wErr := httptest.NewRecorder()
	cErr, _ := gin.CreateTestContext(wErr)
	JSONError(cErr, http.StatusBadRequest, "operation failed", errors.New("custom error detail"))

	assert.Equal(t, http.StatusBadRequest, wErr.Code)
	var respErr StandardResponse
	err = json.Unmarshal(wErr.Body.Bytes(), &respErr)
	assert.NoError(t, err)
	assert.False(t, respErr.Success)
	assert.Equal(t, "operation failed", respErr.Message)
	assert.Equal(t, "custom error detail", respErr.Error)

	// Test JSONError with string detail
	wErrStr := httptest.NewRecorder()
	cErrStr, _ := gin.CreateTestContext(wErrStr)
	JSONError(cErrStr, http.StatusInternalServerError, "server failure", "string error detail")

	assert.Equal(t, http.StatusInternalServerError, wErrStr.Code)
}
