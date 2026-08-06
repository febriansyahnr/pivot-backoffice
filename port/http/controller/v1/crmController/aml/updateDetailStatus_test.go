package aml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMAmlController_UpdateDetailStatus(t *testing.T) {
	testCases := []struct {
		name               string
		profileID          string
		merchantID         string
		requestBody        amlcommon.UpdateDetailStatusRequest
		mockSetup          func(*mocks.IAmlService)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:       "SUCCESS: update status to DISMISS",
			profileID:  "e_tr_wci_123456",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("UpdateDetailStatusByProfileId",
					mock.Anything,
					"e_tr_wci_123456",
					"merchant-123",
					mock.AnythingOfType("*amlcommon.UpdateDetailStatusRequest")).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"message": "Status updated successfully",
			},
		},
		{
			name:       "SUCCESS: update status to PENDING",
			profileID:  "e_tr_wci_567890",
			merchantID: "merchant-789",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "Test User",
				DOB:    "1992-03-10",
				Status: amlcommon.DetailStatusPending,
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("UpdateDetailStatusByProfileId",
					mock.Anything,
					"e_tr_wci_567890",
					"merchant-789",
					mock.AnythingOfType("*amlcommon.UpdateDetailStatusRequest")).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"message": "Status updated successfully",
			},
		},
		{
			name:       "SUCCESS: update status to ON_MONITOR",
			profileID:  "e_tr_wci_789012",
			merchantID: "merchant-456",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "Jane Smith",
				DOB:    "1985-05-15",
				Status: amlcommon.DetailStatusOnMonitor,
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("UpdateDetailStatusByProfileId",
					mock.Anything,
					"e_tr_wci_789012",
					"merchant-456",
					mock.AnythingOfType("*amlcommon.UpdateDetailStatusRequest")).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"message": "Status updated successfully",
			},
		},
		{
			name:       "FAILED: missing merchantId query parameter",
			profileID:  "e_tr_wci_123456",
			merchantID: "",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: empty profileId path parameter",
			profileID:  "",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: invalid request body - missing name",
			profileID:  "e_tr_wci_123456",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: invalid request body - missing dob",
			profileID:  "e_tr_wci_123456",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: invalid request body - missing status",
			profileID:  "e_tr_wci_123456",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name: "John Doe",
				DOB:  "1990-01-01",
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: invalid status value",
			profileID:  "e_tr_wci_123456",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: "INVALID_STATUS",
			},
			mockSetup:          func(mockService *mocks.IAmlService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: service returns error - merchant not found",
			profileID:  "e_tr_wci_123456",
			merchantID: "nonexistent-merchant",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("UpdateDetailStatusByProfileId",
					mock.Anything,
					"e_tr_wci_123456",
					"nonexistent-merchant",
					mock.AnythingOfType("*amlcommon.UpdateDetailStatusRequest")).Return(constant.ErrMerchantNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "FAILED: service returns error - data not found",
			profileID:  "e_tr_wci_999999",
			merchantID: "merchant-123",
			requestBody: amlcommon.UpdateDetailStatusRequest{
				Name:   "Nonexistent User",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockService *mocks.IAmlService) {
				mockService.On("UpdateDetailStatusByProfileId",
					mock.Anything,
					"e_tr_wci_999999",
					"merchant-123",
					mock.AnythingOfType("*amlcommon.UpdateDetailStatusRequest")).Return(constant.ErrDataNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockService := mocks.NewIAmlService(t)
			tc.mockSetup(mockService)

			// Create controller
			validate := validator.New()
			controller := &CRMAmlController{
				amlService: mockService,
				validate:   validate,
			}

			// Create request body
			requestBody, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/crm/v1/aml/screening/"+tc.profileID+"/status?merchantId="+tc.merchantID, bytes.NewReader(requestBody))

			// Setup Chi router context for path parameters
			rctx := chi.NewRouteContext()
			if tc.profileID != "" {
				rctx.URLParams.Add("profileId", tc.profileID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			controller.UpdateDetailStatus(w, req)

			// Assertions
			assert.Equal(t, tc.expectedStatusCode, w.Code)

			if tc.expectedStatusCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				data, exists := response["data"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tc.expectedResponse["message"], data["message"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestCRMAmlController_UpdateDetailStatus_InvalidJSON(t *testing.T) {
	// Setup mocks
	mockService := mocks.NewIAmlService(t)

	// Create controller
	validate := validator.New()
	controller := &CRMAmlController{
		amlService: mockService,
		validate:   validate,
	}

	// Create request with invalid JSON
	invalidJSON := `{"name": "John Doe", "dob": "1990-01-01", "status": }`
	req := httptest.NewRequest(http.MethodPut, "/crm/v1/aml/screening/e_tr_wci_123456/status?merchantId=merchant-123", bytes.NewReader([]byte(invalidJSON)))

	// Setup Chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("profileId", "e_tr_wci_123456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call the handler
	controller.UpdateDetailStatus(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no service calls were made
	mockService.AssertExpectations(t)
}