package aml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMAmlController_Screening(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    url.Values
		requestBody    interface{}
		mockSetup      func(*mocks.IAmlService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "SUCCESS: valid screening request",
			queryParams: url.Values{
				"provider":   []string{constant.PROVIDER_ADVANCE_AI},
				"merchantId": []string{"merchant123"},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "John Doe",
				ReferenceID: "ref123",
				DOB:         "1990-01-15",
				Nationality: "ID",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "merchant123").
					Return(&amlcommon.ScreeningResponse{
						Status:        "success",
						TransactionID: "txn123",
						ReferenceID:   "ref123",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "SUCCESS: valid screening request without merchant ID",
			queryParams: url.Values{
				"provider": []string{constant.PROVIDER_ADVANCE_AI},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "Jane Smith",
				ReferenceID: "ref456",
				DOB:         "1985-05-20",
				Nationality: "US",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "").
					Return(&amlcommon.ScreeningResponse{
						Status:        "success",
						TransactionID: "txn456",
						ReferenceID:   "ref456",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "FAIL: missing provider query parameter",
			queryParams: url.Values{},
			requestBody: amlcommon.CheckRequest{
				Name:        "John Doe",
				ReferenceID: "ref123",
				DOB:         "1990-01-15",
				Nationality: "ID",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				// No mock setup needed as this will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "FAIL: invalid JSON request body",
			queryParams: url.Values{
				"provider": []string{constant.PROVIDER_ADVANCE_AI},
			},
			requestBody:    "{invalid json",
			mockSetup:      func(mockService *mocks.IAmlService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "FAIL: validation error - missing required name field",
			queryParams: url.Values{
				"provider": []string{constant.PROVIDER_ADVANCE_AI},
			},
			requestBody: amlcommon.CheckRequest{
				// Missing required Name field - validation should fail
				ReferenceID: "ref123",
			},
			mockSetup:      func(mockService *mocks.IAmlService) {
				// No mock setup needed as validation will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "FAIL: service returns error",
			queryParams: url.Values{
				"provider":   []string{constant.PROVIDER_ADVANCE_AI},
				"merchantId": []string{"merchant123"},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "John Doe",
				ReferenceID: "ref123",
				DOB:         "1990-01-15",
				Nationality: "ID",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "merchant123").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockAmlService := mocks.NewIAmlService(t)
			tc.mockSetup(mockAmlService)

			validate := validator.New()
			controller := New(mockAmlService, validate)

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
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/aml/screening", bytes.NewReader(reqBody))
			req.URL.RawQuery = tc.queryParams.Encode()
			req = req.WithContext(context.Background())

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			controller.Screening(w, req)

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
			mockAmlService.AssertExpectations(t)
		})
	}
}

func TestCRMAmlController_Screening_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    url.Values
		requestBody    interface{}
		mockSetup      func(*mocks.IAmlService)
		expectedStatus int
	}{
		{
			name: "SUCCESS: empty provider value in query",
			queryParams: url.Values{
				"provider": []string{""},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "John Doe",
				ReferenceID: "ref123",
				DOB:         "1990-01-15",
				Nationality: "ID",
			},
			mockSetup:      func(mockService *mocks.IAmlService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SUCCESS: special characters in merchant ID",
			queryParams: url.Values{
				"provider":   []string{constant.PROVIDER_ADVANCE_AI},
				"merchantId": []string{"merchant-123_test"},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "John Doe",
				ReferenceID: "ref123",
				DOB:         "1990-01-15",
				Nationality: "ID",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "merchant-123_test").
					Return(&amlcommon.ScreeningResponse{
						Status:        "success",
						TransactionID: "txn123",
						ReferenceID:   "ref123",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: complex request payload",
			queryParams: url.Values{
				"provider": []string{constant.PROVIDER_ADVANCE_AI},
			},
			requestBody: amlcommon.CheckRequest{
				Name:              "José María García-López",
				ReferenceID:       "ref789",
				DOB:               "1975-12-31",
				Nationality:       "ES",
				Gender:            "M",
				CountryLocation:   "Spain",
				PlaceOfBirth:      "Madrid",
				RegisteredCountry: "ES",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "").
					Return(&amlcommon.ScreeningResponse{
						Status:        "success",
						TransactionID: "txn789",
						ReferenceID:   "ref789",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: minimal required fields only",
			queryParams: url.Values{
				"provider": []string{constant.PROVIDER_ADVANCE_AI},
			},
			requestBody: amlcommon.CheckRequest{
				Name:        "Test User",
				ReferenceID: "minimal123",
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("Screening", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest"), constant.PROVIDER_ADVANCE_AI, "").
					Return(&amlcommon.ScreeningResponse{
						Status:        "success",
						TransactionID: "txnMinimal",
						ReferenceID:   "minimal123",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockAmlService := mocks.NewIAmlService(t)
			tc.mockSetup(mockAmlService)

			validate := validator.New()
			controller := New(mockAmlService, validate)

			// Prepare request body
			reqBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/aml/screening", bytes.NewReader(reqBody))
			req.URL.RawQuery = tc.queryParams.Encode()
			req = req.WithContext(context.Background())

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute
			controller.Screening(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Verify all expectations were met
			mockAmlService.AssertExpectations(t)
		})
	}
}

func TestNew(t *testing.T) {
	mockAmlService := mocks.NewIAmlService(t)
	validate := validator.New()

	controller := New(mockAmlService, validate)

	assert.NotNil(t, controller)
	assert.IsType(t, &CRMAmlController{}, controller)

	crmController, ok := controller.(*CRMAmlController)
	assert.True(t, ok)
	assert.Equal(t, mockAmlService, crmController.amlService)
	assert.Equal(t, validate, crmController.validate)
}

