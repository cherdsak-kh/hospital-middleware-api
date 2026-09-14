package usecase

import (
	"log"
	"strings"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/google/uuid"
)

type patientUseCase struct {
	patientRepo domain.PatientRepository
	hisClient   domain.HISClient
}

// NewPatientUseCase creates a new patient search usecase
func NewPatientUseCase(patientRepo domain.PatientRepository, hisClient domain.HISClient) domain.PatientUseCase {
	return &patientUseCase{
		patientRepo: patientRepo,
		hisClient:   hisClient,
	}
}

// SearchPatients searches patients in local database and integrates external HIS if applicable
func (u *patientUseCase) SearchPatients(hospitalID uuid.UUID, query *domain.PatientSearchQuery) ([]domain.PatientResponse, error) {
	// 1. Search local database under the authenticated hospital_id
	patients, err := u.patientRepo.Search(hospitalID, query)
	if err != nil {
		return nil, err
	}

	// Track existing national_id and passport_id to prevent duplicates
	existingNationalIDs := make(map[string]bool)
	existingPassportIDs := make(map[string]bool)
	for _, p := range patients {
		if p.NationalID != "" {
			existingNationalIDs[p.NationalID] = true
		}
		if p.PassportID != "" {
			existingPassportIDs[p.PassportID] = true
		}
	}

	// 2. If search criteria includes national_id or passport_id, check external HIS Client
	if u.hisClient != nil && query != nil {
		var searchID string
		if strings.TrimSpace(query.NationalID) != "" {
			searchID = strings.TrimSpace(query.NationalID)
		} else if strings.TrimSpace(query.PassportID) != "" {
			searchID = strings.TrimSpace(query.PassportID)
		}

		if searchID != "" && !existingNationalIDs[searchID] && !existingPassportIDs[searchID] {
			hisResp, err := u.hisClient.SearchPatient(searchID)
			if err != nil {
				// Log external HIS communication failure without breaking the API response
				log.Printf("External HIS client error for ID %s: %v\n", searchID, err)
			} else if hisResp != nil {
				// Check if the external patient is already in our list
				alreadyInList := false
				if hisResp.NationalID != "" && existingNationalIDs[hisResp.NationalID] {
					alreadyInList = true
				}
				if hisResp.PassportID != "" && existingPassportIDs[hisResp.PassportID] {
					alreadyInList = true
				}

				if !alreadyInList {
					// Persist newly discovered patient under this hospital for data isolation
					newPatient := domain.Patient{
						ID:           uuid.New(),
						HospitalID:   hospitalID,
						PatientHN:    hisResp.PatientHN,
						NationalID:   hisResp.NationalID,
						PassportID:   hisResp.PassportID,
						FirstNameTH:  hisResp.FirstNameTH,
						MiddleNameTH: hisResp.MiddleNameTH,
						LastNameTH:   hisResp.LastNameTH,
						FirstNameEN:  hisResp.FirstNameEN,
						MiddleNameEN: hisResp.MiddleNameEN,
						LastNameEN:   hisResp.LastNameEN,
						DateOfBirth:  hisResp.DateOfBirth,
						Gender:       hisResp.Gender,
						PhoneNumber:  hisResp.PhoneNumber,
						Email:        hisResp.Email,
					}

					if saveErr := u.patientRepo.Create(&newPatient); saveErr == nil {
						patients = append(patients, newPatient)
					} else {
						log.Printf("Warning: failed to persist HIS patient record: %v\n", saveErr)
						patients = append(patients, newPatient)
					}
				}
			}
		}
	}

	// 3. Format into standardized responses
	responses := make([]domain.PatientResponse, 0, len(patients))
	for _, p := range patients {
		responses = append(responses, domain.PatientResponse{
			ID:           p.ID,
			HospitalID:   p.HospitalID,
			PatientHN:    p.PatientHN,
			NationalID:   p.NationalID,
			PassportID:   p.PassportID,
			FirstNameTH:  p.FirstNameTH,
			MiddleNameTH: p.MiddleNameTH,
			LastNameTH:   p.LastNameTH,
			FirstNameEN:  p.FirstNameEN,
			MiddleNameEN: p.MiddleNameEN,
			LastNameEN:   p.LastNameEN,
			DateOfBirth:  p.DateOfBirth,
			Gender:       p.Gender,
			PhoneNumber:  p.PhoneNumber,
			Email:        p.Email,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		})
	}

	return responses, nil
}
