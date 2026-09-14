package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/cherdsak-kh/hospital-middleware-api/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrUsernameTaken    = errors.New("username is already taken")
	ErrHospitalNotFound = errors.New("hospital not found")
	ErrInvalidAuth      = errors.New("invalid username or password")
)

type staffUseCase struct {
	staffRepo      domain.StaffRepository
	hospitalRepo   domain.HospitalRepository
	jwtSecret      string
	jwtExpiryHours int
}

// NewStaffUseCase creates a new staff business logic usecase
func NewStaffUseCase(
	staffRepo domain.StaffRepository,
	hospitalRepo domain.HospitalRepository,
	jwtSecret string,
	jwtExpiryHours int,
) domain.StaffUseCase {
	if jwtExpiryHours <= 0 {
		jwtExpiryHours = 24
	}
	return &staffUseCase{
		staffRepo:      staffRepo,
		hospitalRepo:   hospitalRepo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

// CreateStaff handles staff registration with password hashing and hospital association
func (u *staffUseCase) CreateStaff(req *domain.StaffCreateRequest) (*domain.Staff, error) {
	cleanUsername := strings.TrimSpace(req.Username)
	if cleanUsername == "" {
		return nil, errors.New("username cannot be empty")
	}

	// Check if username already exists across the system
	existingStaff, err := u.staffRepo.FindByUsername(cleanUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing staff: %w", err)
	}
	if existingStaff != nil {
		return nil, ErrUsernameTaken
	}

	// Find or automatically create hospital if it does not exist yet
	hospital, err := u.hospitalRepo.FindOrCreate(strings.TrimSpace(req.Hospital))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hospital: %w", err)
	}

	// Hash password with bcrypt
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newStaff := &domain.Staff{
		ID:           uuid.New(),
		HospitalID:   hospital.ID,
		Hospital:     *hospital,
		Username:     cleanUsername,
		PasswordHash: hashedPassword,
		FullName:     strings.TrimSpace(req.FullName),
	}

	if err := u.staffRepo.Create(newStaff); err != nil {
		return nil, fmt.Errorf("failed to create staff record: %w", err)
	}

	return newStaff, nil
}

// LoginStaff verifies credentials within the specified hospital and issues a signed JWT
func (u *staffUseCase) LoginStaff(req *domain.StaffLoginRequest) (*domain.StaffLoginResponse, error) {
	cleanUsername := strings.TrimSpace(req.Username)
	cleanHospital := strings.TrimSpace(req.Hospital)

	// Resolve hospital
	hospital, err := u.hospitalRepo.FindByCodeOrName(cleanHospital)
	if err != nil {
		return nil, fmt.Errorf("failed to find hospital: %w", err)
	}
	if hospital == nil {
		return nil, ErrHospitalNotFound
	}

	// Look up staff under this specific hospital
	staff, err := u.staffRepo.FindByUsernameAndHospital(cleanUsername, hospital.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find staff: %w", err)
	}
	if staff == nil {
		return nil, ErrInvalidAuth
	}

	// Check bcrypt password
	if !utils.CheckPassword(staff.PasswordHash, req.Password) {
		return nil, ErrInvalidAuth
	}

	// Generate JWT token containing staff_id and hospital_id claims
	token, err := utils.GenerateToken(
		staff.ID,
		hospital.ID,
		staff.Username,
		u.jwtSecret,
		u.jwtExpiryHours,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &domain.StaffLoginResponse{
		Token:        token,
		StaffID:      staff.ID,
		Username:     staff.Username,
		HospitalID:   hospital.ID,
		HospitalName: hospital.Name,
	}, nil
}
