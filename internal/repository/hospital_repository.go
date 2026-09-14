package repository

import (
	"errors"
	"strings"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type hospitalRepository struct {
	db *gorm.DB
}

// NewHospitalRepository creates a new hospital repository instance
func NewHospitalRepository(db *gorm.DB) domain.HospitalRepository {
	return &hospitalRepository{db: db}
}

func (r *hospitalRepository) FindByID(id uuid.UUID) (*domain.Hospital, error) {
	var hospital domain.Hospital
	err := r.db.Where("id = ?", id).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) FindByCode(code string) (*domain.Hospital, error) {
	var hospital domain.Hospital
	err := r.db.Where("LOWER(code) = LOWER(?)", strings.TrimSpace(code)).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) FindByName(name string) (*domain.Hospital, error) {
	var hospital domain.Hospital
	err := r.db.Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) FindByCodeOrName(codeOrName string) (*domain.Hospital, error) {
	trimmed := strings.TrimSpace(codeOrName)

	// Check if valid UUID first
	if parsedUUID, err := uuid.Parse(trimmed); err == nil {
		h, err := r.FindByID(parsedUUID)
		if err == nil && h != nil {
			return h, nil
		}
	}

	// Search by code or name
	var hospital domain.Hospital
	err := r.db.Where("LOWER(code) = LOWER(?) OR LOWER(name) = LOWER(?)", trimmed, trimmed).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hospital, nil
}

func (r *hospitalRepository) Create(hospital *domain.Hospital) error {
	return r.db.Create(hospital).Error
}

func (r *hospitalRepository) FindOrCreate(identifier string) (*domain.Hospital, error) {
	existing, err := r.FindByCodeOrName(identifier)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Generate code from identifier
	cleanCode := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(identifier), " ", "_"))
	newHospital := &domain.Hospital{
		ID:   uuid.New(),
		Code: cleanCode,
		Name: strings.TrimSpace(identifier),
	}

	if err := r.Create(newHospital); err != nil {
		return nil, err
	}

	return newHospital, nil
}
