package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
)

type hisClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHISClient creates an HTTP client for Hospital A HIS API
func NewHISClient(baseURL string, timeout time.Duration) domain.HISClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	cleanBaseURL := strings.TrimRight(baseURL, "/")

	return &hisClient{
		baseURL: cleanBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SearchPatient fetches patient data from Hospital A HIS by national_id or passport_id
func (c *hisClient) SearchPatient(id string) (*domain.HospitalAPatientResponse, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil, nil
	}

	targetURL := fmt.Sprintf("%s/patient/search/%s", c.baseURL, url.PathEscape(trimmedID))

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HIS request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hospital-middleware-api/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HIS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HIS returned unexpected status code: %d", resp.StatusCode)
	}

	var patientResp domain.HospitalAPatientResponse
	if err := json.NewDecoder(resp.Body).Decode(&patientResp); err != nil {
		return nil, fmt.Errorf("failed to decode HIS response: %w", err)
	}

	return &patientResp, nil
}
