package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigAndGetDSN(t *testing.T) {
	// Set custom environment variables for test
	os.Setenv("PORT", "5001")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "agnos_user")
	os.Setenv("DB_PASSWORD", "agnos_pass")
	os.Setenv("DB_NAME", "hospital_agnos")
	os.Setenv("JWT_SECRET", "super_secret_agnos")
	os.Setenv("JWT_EXPIRY_HOURS", "12")
	os.Setenv("APP_ENV", "testing")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "5001", cfg.Port)
	assert.Equal(t, "db.example.com", cfg.DBHost)
	assert.Equal(t, "agnos_user", cfg.DBUser)
	assert.Equal(t, 12, cfg.JWTExpiryHours)
	assert.Equal(t, "testing", cfg.AppEnv)

	dsn := cfg.GetDSN()
	assert.Contains(t, dsn, "host=db.example.com")
	assert.Contains(t, dsn, "user=agnos_user")
	assert.Contains(t, dsn, "dbname=hospital_agnos")
}
