package domain

import (
	"time"

	"github.com/google/uuid"
)

// Patient represents a patient entity in the hospital database
type Patient struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalID   uuid.UUID `gorm:"type:uuid;not null;index" json:"hospital_id"`
	Hospital     Hospital  `gorm:"foreignKey:HospitalID;constraint:OnDelete:CASCADE" json:"hospital,omitempty"`
	PatientHN    string    `gorm:"type:varchar(100);index" json:"patient_hn"`
	NationalID   string    `gorm:"type:varchar(20);index" json:"national_id"`
	PassportID   string    `gorm:"type:varchar(50);index" json:"passport_id"`
	FirstNameTH  string    `gorm:"type:varchar(100);index" json:"first_name_th"`
	MiddleNameTH string    `gorm:"type:varchar(100)" json:"middle_name_th"`
	LastNameTH   string    `gorm:"type:varchar(100);index" json:"last_name_th"`
	FirstNameEN  string    `gorm:"type:varchar(100);index" json:"first_name_en"`
	MiddleNameEN string    `gorm:"type:varchar(100)" json:"middle_name_en"`
	LastNameEN   string    `gorm:"type:varchar(100);index" json:"last_name_en"`
	DateOfBirth  string    `gorm:"type:varchar(20)" json:"date_of_birth"`
	Gender       string    `gorm:"type:varchar(1)" json:"gender"`
	PhoneNumber  string    `gorm:"type:varchar(50);index" json:"phone_number"`
	Email        string    `gorm:"type:varchar(255);index" json:"email"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// PatientSearchQuery holds optional search criteria for patients
type PatientSearchQuery struct {
	NationalID   string `form:"national_id" json:"national_id"`
	PassportID   string `form:"passport_id" json:"passport_id"`
	FirstName    string `form:"first_name" json:"first_name"`
	MiddleName   string `form:"middle_name" json:"middle_name"`
	LastName     string `form:"last_name" json:"last_name"`
	DateOfBirth  string `form:"date_of_birth" json:"date_of_birth"`
	PhoneNumber  string `form:"phone_number" json:"phone_number"`
	Email        string `form:"email" json:"email"`
	PatientHN    string `form:"patient_hn" json:"patient_hn"`
	FirstNameTH  string `form:"first_name_th" json:"first_name_th"`
	LastNameTH   string `form:"last_name_th" json:"last_name_th"`
	FirstNameEN  string `form:"first_name_en" json:"first_name_en"`
	LastNameEN   string `form:"last_name_en" json:"last_name_en"`
}

// PatientResponse represents standardized patient output
type PatientResponse struct {
	ID           uuid.UUID `json:"id"`
	HospitalID   uuid.UUID `json:"hospital_id"`
	PatientHN    string    `json:"patient_hn"`
	NationalID   string    `json:"national_id"`
	PassportID   string    `json:"passport_id"`
	FirstNameTH  string    `json:"first_name_th"`
	MiddleNameTH string    `json:"middle_name_th"`
	LastNameTH   string    `json:"last_name_th"`
	FirstNameEN  string    `json:"first_name_en"`
	MiddleNameEN string    `json:"middle_name_en"`
	LastNameEN   string    `json:"last_name_en"`
	DateOfBirth  string    `json:"date_of_birth"`
	Gender       string    `json:"gender"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// HospitalAPatientResponse represents response schema from Hospital A HIS API
type HospitalAPatientResponse struct {
	FirstNameTH  string `json:"first_name_th"`
	MiddleNameTH string `json:"middle_name_th"`
	LastNameTH   string `json:"last_name_th"`
	FirstNameEN  string `json:"first_name_en"`
	MiddleNameEN string `json:"middle_name_en"`
	LastNameEN   string `json:"last_name_en"`
	DateOfBirth  string `json:"date_of_birth"`
	PatientHN    string `json:"patient_hn"`
	NationalID   string `json:"national_id"`
	PassportID   string `json:"passport_id"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Gender       string `json:"gender"`
}

// PatientRepository defines data access methods for patients
type PatientRepository interface {
	Search(hospitalID uuid.UUID, query *PatientSearchQuery) ([]Patient, error)
	Create(patient *Patient) error
	FindByNationalID(hospitalID uuid.UUID, nationalID string) (*Patient, error)
	FindByPassportID(hospitalID uuid.UUID, passportID string) (*Patient, error)
	FindByHN(hospitalID uuid.UUID, hn string) (*Patient, error)
}

// HISClient defines external HIS integration methods
type HISClient interface {
	SearchPatient(id string) (*HospitalAPatientResponse, error)
}

// PatientUseCase defines business logic operations for patient search
type PatientUseCase interface {
	SearchPatients(hospitalID uuid.UUID, query *PatientSearchQuery) ([]PatientResponse, error)
}
