package utils

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test_secret_key_12345"
	staffID := uuid.New()
	hospitalID := uuid.New()
	username := "doctor_somchai"

	// Positive test: Generate and validate token
	token, err := GenerateToken(staffID, hospitalID, username, secret, 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(token, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, staffID, claims.StaffID)
	assert.Equal(t, hospitalID, claims.HospitalID)
	assert.Equal(t, username, claims.Username)

	// Negative test: Invalid secret key
	_, errWrongSecret := ValidateToken(token, "wrong_secret_key")
	assert.Error(t, errWrongSecret)

	// Negative test: Malformed token
	_, errMalformed := ValidateToken("invalid.token.structure", secret)
	assert.Error(t, errMalformed)

	// Negative test: Expired token
	expiredToken, err := GenerateToken(staffID, hospitalID, username, secret, -2)
	assert.NoError(t, err)
	_, errExpired := ValidateToken(expiredToken, secret)
	assert.Error(t, errExpired)

	// Negative test: Signing algorithm not HMAC (SigningMethodNone)
	noneTokenObj := jwt.NewWithClaims(jwt.SigningMethodNone, &JWTClaims{
		StaffID:  staffID,
		Username: username,
	})
	noneTokenString, _ := noneTokenObj.SignedString(jwt.UnsafeAllowNoneSignatureType)
	_, errNone := ValidateToken(noneTokenString, secret)
	assert.Error(t, errNone)

	// Negative test: Empty secret in GenerateToken
	_, errEmptyGen := GenerateToken(staffID, hospitalID, username, "   ", 1)
	assert.Error(t, errEmptyGen)

	// Negative test: Empty secret in ValidateToken
	_, errEmptyVal := ValidateToken(token, "   ")
	assert.Error(t, errEmptyVal)
}
