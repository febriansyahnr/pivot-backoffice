package merchant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBlockMerchant(t *testing.T) {
	validMerchantID := uuid.NewString()

	successResponse := &merchantModel.BlockMerchantResponse{
		BlockedMerchantDetails: merchantModel.BlockedMerchantDetails{
			MerchantId:   validMerchantID,
			MerchantName: "Test Merchant",
			BlockedAt:    time.Now(),
		},
		SubMerchants: []merchantModel.BlockedMerchantDetails{},
	}

	testCases := []struct {
		name           string
		merchantID     string
		setupMocks     func(*mockSvc.IMerchantService)
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "SUCCESS: Valid merchant ID",
			merchantID: validMerchantID,
			setupMocks: func(mockMerchantSvc *mockSvc.IMerchantService) {
				mockMerchantSvc.On("BlockMerchant", mock.Anything, validMerchantID).
					Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:       "ERROR: Invalid merchant ID",
			merchantID: "invalid-uuid",
			setupMocks: func(mockMerchantSvc *mockSvc.IMerchantService) {
				// No expectations - should fail before calling service
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:       "ERROR: Empty merchant ID",
			merchantID: "",
			setupMocks: func(mockMerchantSvc *mockSvc.IMerchantService) {
				// No expectations - should fail before calling service
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockMerchantSvc := mockSvc.NewIMerchantService(t)
			mockValidator := validator.New()

			tc.setupMocks(mockMerchantSvc)

			// Create controller
			controller := &CRMMerchantController{
				merchantSvc: mockMerchantSvc,
				validate:    mockValidator,
			}

			// Setup router
			router := chi.NewRouter()
			router.Post("/api/v1/merchants/{id}/block", controller.BlockMerchant)

			// Create request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/merchants/"+tc.merchantID+"/block",
				nil,
			)

			// Create response recorder
			recorder := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(recorder, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, recorder.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			if tc.expectError {
				require.Contains(t, response, "code")
				require.Contains(t, response, "errors")
				assert.NotContains(t, response, "data")
			} else {
				require.Contains(t, response, "data")
				require.Contains(t, response, "code")
				assert.Equal(t, "00", response["code"])

				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "merchantId")
				assert.Contains(t, data, "merchantName")
				assert.Contains(t, data, "blockedAt")
				assert.Contains(t, data, "subMerchants")
			}
		})
	}
}

func TestUnblockMerchant(t *testing.T) {
	validMerchantID := uuid.NewString()

	successResponse := &merchantModel.UnblockMerchantResponse{
		UnblockedMerchantDetails: merchantModel.UnblockedMerchantDetails{
			MerchantId:   validMerchantID,
			MerchantName: "Test Merchant",
			UnblockedAt:  time.Now(),
		},
		SubMerchants: []merchantModel.UnblockedMerchantDetails{},
	}

	testCases := []struct {
		name           string
		merchantID     string
		setupMocks     func(*mockSvc.IMerchantService)
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "SUCCESS: Valid merchant ID",
			merchantID: validMerchantID,
			setupMocks: func(mockMerchantSvc *mockSvc.IMerchantService) {
				mockMerchantSvc.On("UnblockMerchant", mock.Anything, validMerchantID).
					Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "ERROR: Invalid merchant ID",
			merchantID:     "invalid-uuid",
			setupMocks:     func(mockMerchantSvc *mockSvc.IMerchantService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "ERROR: Empty merchant ID",
			merchantID:     "",
			setupMocks:     func(mockMerchantSvc *mockSvc.IMerchantService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:       "ERROR: Service returns error",
			merchantID: validMerchantID,
			setupMocks: func(mockMerchantSvc *mockSvc.IMerchantService) {
				mockMerchantSvc.On("UnblockMerchant", mock.Anything, validMerchantID).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMerchantSvc := mockSvc.NewIMerchantService(t)
			tc.setupMocks(mockMerchantSvc)

			controller := &CRMMerchantController{
				merchantSvc: mockMerchantSvc,
				validate:    validator.New(),
			}

			router := chi.NewRouter()
			router.Post("/api/v1/merchants/{id}/unblock", controller.UnblockMerchant)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants/"+tc.merchantID+"/unblock", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			if tc.expectError {
				require.Contains(t, response, "code")
				require.Contains(t, response, "errors")
				assert.NotContains(t, response, "data")
			} else {
				require.Contains(t, response, "data")
				require.Contains(t, response, "code")
				assert.Equal(t, "00", response["code"])

				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "merchantId")
				assert.Contains(t, data, "merchantName")
				assert.Contains(t, data, "unblockedAt")
				assert.Contains(t, data, "subMerchants")
			}
		})
	}
}
