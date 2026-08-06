package dukcapil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-playground/validator/v10"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMDukcapilController_VerifyIdentity(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    url.Values
		requestBody    interface{}
		mockSetup      func(*mocks.IDukcapilService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SUCCESS: valid verification request with merchantId",
			queryParams: url.Values{
				"merchantId": []string{"merchant123"},
			},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:      "1234567890123456",
				Name:     "John Doe",
				Gender:   "L",
				DOB:      "01-01-1990",
				POB:      "Jakarta",
				Job:      "Engineer",
				Address:  "Jl. Test No 1",
				RT:       "001",
				RW:       "002",
				District: "Jakarta Selatan",
				Province: "DKI Jakarta",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID: "ref123",
						Status:      dukcapilmodel.StatusMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{
							{
								Field:     dukcapilmodel.FieldName,
								Score:     100,
								Threshold: 100,
								Status:    dukcapilmodel.StatusMatched,
							},
							{
								Field:     dukcapilmodel.FieldGender,
								Score:     100,
								Threshold: 100,
								Status:    dukcapilmodel.StatusMatched,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "SUCCESS: valid verification request without merchantId",
			queryParams: url.Values{},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "Jane Smith",
				DOB:  "15-05-1985",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID:  "ref456",
						Status:       dukcapilmodel.StatusNotMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "FAIL: invalid JSON request body",
			queryParams:    url.Values{},
			requestBody:    "{invalid json",
			mockSetup:      func(mockService *mocks.IDukcapilService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "FAIL: validation error - missing required NIK field",
			queryParams: url.Values{},
			requestBody: dukcapilmodel.VerifyRequest{
				// Missing required NIK field - validation should fail
				Name: "John Doe",
				DOB:  "01-01-1990",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				// No mock setup needed as validation will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "FAIL: validation error - missing required FullName field",
			queryParams: url.Values{},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK: "1234567890123456",
				// Missing required FullName field - validation should fail
				DOB: "01-01-1990",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				// No mock setup needed as validation will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "FAIL: validation error - missing required BirthDate field",
			queryParams: url.Values{},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "John Doe",
				// Missing required BirthDate field - validation should fail
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				// No mock setup needed as validation will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "FAIL: service returns error",
			queryParams: url.Values{
				"merchantId": []string{"merchant123"},
			},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "John Doe",
				DOB:  "01-01-1990",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockDukcapilService := mocks.NewIDukcapilService(t)
			tc.mockSetup(mockDukcapilService)

			validate := validator.New()
			controller := New(mockDukcapilService, validate)

			// Prepare request body
			var reqBody []byte
			if str, ok := tc.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				var err error
				reqBody, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/dukcapil/verify-identity", bytes.NewReader(reqBody))
			req.URL.RawQuery = tc.queryParams.Encode()
			req = req.WithContext(context.Background())

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			controller.VerifyIdentity(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				// For successful cases, verify response structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "data")
			}

			// Verify all expectations were met
			mockDukcapilService.AssertExpectations(t)
		})
	}
}

func TestCRMDukcapilController_VerifyIdentity_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    url.Values
		requestBody    interface{}
		mockSetup      func(*mocks.IDukcapilService)
		expectedStatus int
	}{
		{
			name: "SUCCESS: complex request payload with all fields",
			queryParams: url.Values{
				"merchantId": []string{"merchant-complex-123"},
			},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:      "3271012345678901",
				Name:     "José María García-López",
				Gender:   "L",
				DOB:      "31-12-1975",
				POB:      "Jakarta Selatan",
				Job:      "Software Engineer",
				Address:  "Jl. Sudirman Kav. 52-53",
				RT:       "010",
				RW:       "005",
				Village:  "Senayan",
				District: "Kebayoran Baru",
				Regency:  "Jakarta Selatan",
				Province: "DKI Jakarta",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID: "ref-complex-789",
						Status:      dukcapilmodel.StatusMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{
							{
								Field:     dukcapilmodel.FieldName,
								Score:     95,
								Threshold: 90,
								Status:    dukcapilmodel.StatusMatched,
							},
							{
								Field:     dukcapilmodel.FieldAddress,
								Score:     88,
								Threshold: 85,
								Status:    dukcapilmodel.StatusMatched,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "SUCCESS: minimal required fields only",
			queryParams: url.Values{},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "Minimal User",
				DOB:  "01-01-1990",
				// All other fields empty
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID:  "ref-minimal-456",
						Status:       dukcapilmodel.StatusMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: special characters in merchant ID",
			queryParams: url.Values{
				"merchantId": []string{"merchant-123_test"},
			},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "Test User",
				DOB:  "15-08-1985",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID: "ref-special-123",
						Status:      dukcapilmodel.StatusNotMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{
							{
								Field:     dukcapilmodel.FieldName,
								Score:     75,
								Threshold: 90,
								Status:    dukcapilmodel.StatusNotMatched,
							},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: empty merchant ID in query param",
			queryParams: url.Values{
				"merchantId": []string{""},
			},
			requestBody: dukcapilmodel.VerifyRequest{
				NIK:  "1234567890123456",
				Name: "Empty Merchant Test",
				DOB:  "20-06-1990",
			},
			mockSetup: func(mockService *mocks.IDukcapilService) {
				mockService.On("VerifyIdentity", mock.Anything, mock.AnythingOfType("*dukcapilmodel.IdentityVerificationRequest")).
					Return(&dukcapilmodel.IdentityVerificationResponse{
						ReferenceID:  "ref-empty-merchant",
						Status:       dukcapilmodel.StatusMatched,
						FieldResults: []dukcapilmodel.DukcapilFieldResult{},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockDukcapilService := mocks.NewIDukcapilService(t)
			tc.mockSetup(mockDukcapilService)

			validate := validator.New()
			controller := New(mockDukcapilService, validate)

			// Prepare request body
			reqBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/dukcapil/verify-identity", bytes.NewReader(reqBody))
			req.URL.RawQuery = tc.queryParams.Encode()
			req = req.WithContext(context.Background())

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			controller.VerifyIdentity(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Verify all expectations were met
			mockDukcapilService.AssertExpectations(t)
		})
	}
}

func TestNew(t *testing.T) {
	mockDukcapilService := mocks.NewIDukcapilService(t)
	validate := validator.New()

	controller := New(mockDukcapilService, validate)

	assert.NotNil(t, controller)
	assert.IsType(t, &CRMDukcapilController{}, controller)
	assert.Equal(t, mockDukcapilService, controller.dukcapilService)
	assert.Equal(t, validate, controller.validate)
}
