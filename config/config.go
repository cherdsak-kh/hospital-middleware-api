package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application
type Config struct {
	Port             string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	JWTSecret        string
	JWTExpiryHours   int
	HospitalABaseURL string
	AppEnv           string
}

// LoadConfig loads configuration from environment variables and optional .env file
func LoadConfig() (*Config, error) {
	// Load .env if present (ignore error if not found)
	_ = godotenv.Load()

	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		expiryHours = 24
	}

	cfg := &Config{
		Port:             getEnv("PORT", "5000"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "hospital_db"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		JWTSecret:        getEnv("JWT_SECRET", "default_secret_key_change_in_production"),
		JWTExpiryHours:   expiryHours,
		HospitalABaseURL: getEnv("HOSPITAL_A_BASE_URL", "https://hospital-a.api.co.th"),
		AppEnv:           getEnv("APP_ENV", "development"),
	}

	return cfg, nil
}

// GetDSN formats PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Bangkok",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
