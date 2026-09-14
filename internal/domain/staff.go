package domain

import (
	"time"

	"github.com/google/uuid"
)

// Staff represents hospital personnel with access to patient records
type Staff struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalID   uuid.UUID `gorm:"type:uuid;not null;index" json:"hospital_id"`
	Hospital     Hospital  `gorm:"foreignKey:HospitalID;constraint:OnDelete:CASCADE" json:"hospital,omitempty"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName     string    `gorm:"type:varchar(255)" json:"full_name"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// StaffCreateRequest represents payload for creating new staff
type StaffCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Hospital string `json:"hospital" binding:"required"`
	FullName string `json:"full_name"`
}

// StaffLoginRequest represents payload for staff authentication
type StaffLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

// StaffLoginResponse represents token response upon successful login
type StaffLoginResponse struct {
	Token        string    `json:"token"`
	StaffID      uuid.UUID `json:"staff_id"`
	Username     string    `json:"username"`
	HospitalID   uuid.UUID `json:"hospital_id"`
	HospitalName string    `json:"hospital_name,omitempty"`
}

// StaffRepository defines database operations for staff
type StaffRepository interface {
	Create(staff *Staff) error
	FindByID(id uuid.UUID) (*Staff, error)
	FindByUsername(username string) (*Staff, error)
	FindByUsernameAndHospital(username string, hospitalID uuid.UUID) (*Staff, error)
}

// StaffUseCase defines business logic operations for staff
type StaffUseCase interface {
	CreateStaff(req *StaffCreateRequest) (*Staff, error)
	LoginStaff(req *StaffLoginRequest) (*StaffLoginResponse, error)
}
