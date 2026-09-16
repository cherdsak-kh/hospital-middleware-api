package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// JWTClaims holds custom JWT payload information
type JWTClaims struct {
	StaffID    uuid.UUID `json:"staff_id"`
	Username   string    `json:"username"`
	HospitalID uuid.UUID `json:"hospital_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new signed JWT token with staff and hospital metadata
func GenerateToken(staffID, hospitalID uuid.UUID, username, secret string, expiryHours int) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("jwt secret cannot be empty")
	}

	expirationTime := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	claims := &JWTClaims{
		StaffID:    staffID,
		Username:   username,
		HospitalID: hospitalID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "hospital-middleware-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and verifies the signed JWT token
func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jwt secret cannot be empty")
	}

	claims := &JWTClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
