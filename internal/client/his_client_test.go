package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cherdsak-kh/hospital-middleware-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestHISClient_SearchPatient(t *testing.T) {
	// Mock external Hospital A HIS HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/patient/search/1100100111111":
			patient := domain.HospitalAPatientResponse{
				PatientHN:   "HN-A-100",
				NationalID:  "1100100111111",
				FirstNameTH: "ประสิทธิ์",
				LastNameTH:  "แซ่ตั้ง",
				FirstNameEN: "Prasit",
				LastNameEN:  "Saetang",
				PhoneNumber: "0819998888",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(patient)

		case "/patient/search/NOTFOUND":
			w.WriteHeader(http.StatusNotFound)

		case "/patient/search/ERROR500":
			w.WriteHeader(http.StatusInternalServerError)

		case "/patient/search/BADJSON":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json-content}"))

		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer mockServer.Close()

	// Test default timeout fallback when timeout <= 0
	clientDefaultTimeout := NewHISClient(mockServer.URL, 0)
	assert.NotNil(t, clientDefaultTimeout)

	client := NewHISClient(mockServer.URL, 2*time.Second)

	// 1. Positive test: Successfully found patient
	p, err := client.SearchPatient("1100100111111")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "HN-A-100", p.PatientHN)
	assert.Equal(t, "1100100111111", p.NationalID)

	// 2. Negative test: Empty ID
	pEmpty, errEmpty := client.SearchPatient("")
	assert.NoError(t, errEmpty)
	assert.Nil(t, pEmpty)

	// 3. Negative test: Patient not found (404)
	pNotFound, errNotFound := client.SearchPatient("NOTFOUND")
	assert.NoError(t, errNotFound)
	assert.Nil(t, pNotFound)

	// 4. Negative test: Server error (500)
	pErr, err500 := client.SearchPatient("ERROR500")
	assert.Error(t, err500)
	assert.Nil(t, pErr)

	// 5. Negative test: Malformed JSON response
	pBadJSON, errBadJSON := client.SearchPatient("BADJSON")
	assert.Error(t, errBadJSON)
	assert.Nil(t, pBadJSON)
	assert.Contains(t, errBadJSON.Error(), "failed to decode HIS response")

	// 6. Negative test: Network connection failure (closed server)
	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closedServer.URL
	closedServer.Close() // Close immediately to trigger connection refused
	clientClosed := NewHISClient(closedURL, 1*time.Second)
	pConnErr, errConn := clientClosed.SearchPatient("1100100111111")
	assert.Error(t, errConn)
	assert.Nil(t, pConnErr)
	assert.Contains(t, errConn.Error(), "HIS request failed")

	// 7. Negative test: Invalid URL triggering http.NewRequest error
	clientInvalidURL := NewHISClient("http://\x7f-invalid-url", 1*time.Second)
	pReqErr, errReq := clientInvalidURL.SearchPatient("1100100111111")
	assert.Error(t, errReq)
	assert.Nil(t, pReqErr)
	assert.Contains(t, errReq.Error(), "failed to create HIS request")
}
