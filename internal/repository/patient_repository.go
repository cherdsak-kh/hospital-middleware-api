package repository

import (
	"errors"
	"strings"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type patientRepository struct {
	db *gorm.DB
}

// NewPatientRepository creates a new patient repository instance
func NewPatientRepository(db *gorm.DB) domain.PatientRepository {
	return &patientRepository{db: db}
}

// Search queries patients with dynamic criteria, strictly locked to hospital_id for data isolation
func (r *patientRepository) Search(hospitalID uuid.UUID, query *domain.PatientSearchQuery) ([]domain.Patient, error) {
	// Strict Data Isolation: hospital_id condition is ALWAYS enforced
	tx := r.db.Where("hospital_id = ?", hospitalID)

	if query != nil {
		if strings.TrimSpace(query.NationalID) != "" {
			tx = tx.Where("national_id = ?", strings.TrimSpace(query.NationalID))
		}
		if strings.TrimSpace(query.PassportID) != "" {
			tx = tx.Where("passport_id = ?", strings.TrimSpace(query.PassportID))
		}
		if strings.TrimSpace(query.PatientHN) != "" {
			tx = tx.Where("patient_hn = ?", strings.TrimSpace(query.PatientHN))
		}
		if strings.TrimSpace(query.DateOfBirth) != "" {
			tx = tx.Where("date_of_birth = ?", strings.TrimSpace(query.DateOfBirth))
		}
		if strings.TrimSpace(query.PhoneNumber) != "" {
			tx = tx.Where("phone_number = ?", strings.TrimSpace(query.PhoneNumber))
		}
		if strings.TrimSpace(query.Email) != "" {
			tx = tx.Where("LOWER(email) = LOWER(?)", strings.TrimSpace(query.Email))
		}

		// Handle first_name (support both Thai and English matches)
		if strings.TrimSpace(query.FirstName) != "" {
			term := "%" + strings.TrimSpace(query.FirstName) + "%"
			tx = tx.Where("(first_name_th ILIKE ? OR first_name_en ILIKE ?)", term, term)
		} else {
			if strings.TrimSpace(query.FirstNameTH) != "" {
				tx = tx.Where("first_name_th ILIKE ?", "%"+strings.TrimSpace(query.FirstNameTH)+"%")
			}
			if strings.TrimSpace(query.FirstNameEN) != "" {
				tx = tx.Where("first_name_en ILIKE ?", "%"+strings.TrimSpace(query.FirstNameEN)+"%")
			}
		}

		// Handle middle_name
		if strings.TrimSpace(query.MiddleName) != "" {
			term := "%" + strings.TrimSpace(query.MiddleName) + "%"
			tx = tx.Where("(middle_name_th ILIKE ? OR middle_name_en ILIKE ?)", term, term)
		}

		// Handle last_name (support both Thai and English matches)
		if strings.TrimSpace(query.LastName) != "" {
			term := "%" + strings.TrimSpace(query.LastName) + "%"
			tx = tx.Where("(last_name_th ILIKE ? OR last_name_en ILIKE ?)", term, term)
		} else {
			if strings.TrimSpace(query.LastNameTH) != "" {
				tx = tx.Where("last_name_th ILIKE ?", "%"+strings.TrimSpace(query.LastNameTH)+"%")
			}
			if strings.TrimSpace(query.LastNameEN) != "" {
				tx = tx.Where("last_name_en ILIKE ?", "%"+strings.TrimSpace(query.LastNameEN)+"%")
			}
		}
	}

	var patients []domain.Patient
	if err := tx.Order("created_at DESC").Find(&patients).Error; err != nil {
		return nil, err
	}

	return patients, nil
}

func (r *patientRepository) Create(patient *domain.Patient) error {
	return r.db.Create(patient).Error
}

func (r *patientRepository) FindByNationalID(hospitalID uuid.UUID, nationalID string) (*domain.Patient, error) {
	var patient domain.Patient
	err := r.db.Where("hospital_id = ? AND national_id = ?", hospitalID, strings.TrimSpace(nationalID)).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindByPassportID(hospitalID uuid.UUID, passportID string) (*domain.Patient, error) {
	var patient domain.Patient
	err := r.db.Where("hospital_id = ? AND passport_id = ?", hospitalID, strings.TrimSpace(passportID)).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindByHN(hospitalID uuid.UUID, hn string) (*domain.Patient, error) {
	var patient domain.Patient
	err := r.db.Where("hospital_id = ? AND patient_hn = ?", hospitalID, strings.TrimSpace(hn)).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &patient, nil
}
