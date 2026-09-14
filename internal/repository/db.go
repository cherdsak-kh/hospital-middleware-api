package repository

import (
	"log"
	"time"

	"github.com/cherdsak-kh/hospital-middleware-api/config"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes PostgreSQL connection and runs automatic migrations
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ensure uuid-ossp extension is enabled in PostgreSQL
	_ = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error

	// Run AutoMigrate for domain models
	if err := db.AutoMigrate(
		&domain.Hospital{},
		&domain.Staff{},
		&domain.Patient{},
	); err != nil {
		log.Printf("AutoMigrate warning: %v\n", err)
	}

	// Create composite indexes for optimal query performance
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_patients_hospital_national_id ON patients(hospital_id, national_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_patients_hospital_passport_id ON patients(hospital_id, passport_id);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_patients_hospital_phone ON patients(hospital_id, phone_number);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_patients_hospital_email ON patients(hospital_id, email);`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_patients_hospital_name_th ON patients(hospital_id, first_name_th, last_name_th);`).Error

	return db, nil
}
