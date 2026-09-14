package domain

import (
	"time"

	"github.com/google/uuid"
)

// Hospital represents the hospital entity in the database
type Hospital struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// HospitalRepository defines database operations for hospitals
type HospitalRepository interface {
	FindByID(id uuid.UUID) (*Hospital, error)
	FindByCode(code string) (*Hospital, error)
	FindByName(name string) (*Hospital, error)
	FindByCodeOrName(codeOrName string) (*Hospital, error)
	Create(hospital *Hospital) error
	FindOrCreate(identifier string) (*Hospital, error)
}
