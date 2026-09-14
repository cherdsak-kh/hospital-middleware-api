package repository

import (
	"errors"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type staffRepository struct {
	db *gorm.DB
}

// NewStaffRepository creates a new staff repository instance
func NewStaffRepository(db *gorm.DB) domain.StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(staff *domain.Staff) error {
	return r.db.Create(staff).Error
}

func (r *staffRepository) FindByID(id uuid.UUID) (*domain.Staff, error) {
	var staff domain.Staff
	err := r.db.Preload("Hospital").Where("id = ?", id).First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) FindByUsername(username string) (*domain.Staff, error) {
	var staff domain.Staff
	err := r.db.Preload("Hospital").Where("username = ?", username).First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) FindByUsernameAndHospital(username string, hospitalID uuid.UUID) (*domain.Staff, error) {
	var staff domain.Staff
	err := r.db.Preload("Hospital").Where("username = ? AND hospital_id = ?", username, hospitalID).First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}
