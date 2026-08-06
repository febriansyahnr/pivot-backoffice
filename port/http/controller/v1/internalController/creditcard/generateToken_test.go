package creditcard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgMonitor "github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGeneratePaymentToken(t *testing.T) {
	paymentID := uuid.New().String()
	validExpiryTime := time.Now().Add(24 * time.Hour)
	validExpiryAt := validExpiryTime.Format(time.RFC3339)
	generatedToken := "test-payment-token-" + uuid.New().String()

	validPayload := unifiedPaymentModel.GeneratePaymentTokenRequest{
		PaymentID: paymentID,
		ExpiryAt:  validExpiryAt,
	}

	config := &config.Config{}

	testCases := []struct {
		name           string
		mockSetup      func(paymentSvc *mocks.IPaymentService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "SUCCESS: Generate Payment Token",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(nil)
				paymentSvc.
					On("GeneratePaymentToken",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID,
						mock.AnythingOfType("time.Time")).
					Return(generatedToken, nil)
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var apiResponse struct {
					Code    string                                           `json:"code"`
					Message string                                           `json:"message"`
					Data    unifiedPaymentModel.GeneratePaymentTokenResponse `json:"data"`
				}
				err := json.NewDecoder(recorder.Body).Decode(&apiResponse)
				assert.NoError(t, err)
				assert.Equal(t, generatedToken, apiResponse.Data.Token)
			},
		},
		{
			name:      "ERROR: Invalid JSON",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name:      "ERROR: Validation Failure - Missing PaymentID",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: "",
					ExpiryAt:  validExpiryAt,
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name:      "ERROR: Validation Failure - Missing ExpiryAt",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: paymentID,
					ExpiryAt:  "",
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name:      "ERROR: Invalid ExpiryAt Format",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: paymentID,
					ExpiryAt:  "2024-13-45 25:61:61", // Invalid date format
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name:      "ERROR: Invalid ExpiryAt - Not RFC3339",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: paymentID,
					ExpiryAt:  "2024-10-18 10:00:00", // Not RFC3339 format
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name: "ERROR: HandleStrictExpiry Error",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(errors.New("strict expiry handling failed"))
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			validateResp:   nil,
		},
		{
			name: "ERROR: Service Error - Payment Not Found",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(nil)
				paymentSvc.
					On("GeneratePaymentToken",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID,
						mock.AnythingOfType("time.Time")).
					Return("", errors.New("payment not found"))
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp:   nil,
		},
		{
			name: "ERROR: Service Error - Database Error",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(nil)
				paymentSvc.
					On("GeneratePaymentToken",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID,
						mock.AnythingOfType("time.Time")).
					Return("", errors.New("database connection error"))
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp:   nil,
		},
		{
			name: "SUCCESS: Generate Token With Past Expiry (Edge Case)",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(nil)
				paymentSvc.
					On("GeneratePaymentToken",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID,
						mock.AnythingOfType("time.Time")).
					Return(generatedToken, nil)
			},
			setupBody: func(t *testing.T) []byte {
				pastTime := time.Now().Add(-24 * time.Hour)
				payload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: paymentID,
					ExpiryAt:  pastTime.Format(time.RFC3339),
				}
				jsonPayload, err := json.Marshal(payload)
				assert.NoError(t, err)
				return jsonPayload
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var apiResponse struct {
					Code    string                                           `json:"code"`
					Message string                                           `json:"message"`
					Data    unifiedPaymentModel.GeneratePaymentTokenResponse `json:"data"`
				}
				err := json.NewDecoder(recorder.Body).Decode(&apiResponse)
				assert.NoError(t, err)
				assert.Equal(t, generatedToken, apiResponse.Data.Token)
			},
		},
		{
			name: "SUCCESS: Generate Token With Far Future Expiry",
			mockSetup: func(paymentSvc *mocks.IPaymentService) {
				paymentSvc.
					On("HandleStrictExpiry",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID).
					Return(nil)
				paymentSvc.
					On("GeneratePaymentToken",
						mock.AnythingOfType("*context.valueCtx"),
						paymentID,
						mock.AnythingOfType("time.Time")).
					Return(generatedToken, nil)
			},
			setupBody: func(t *testing.T) []byte {
				futureTime := time.Now().Add(365 * 24 * time.Hour) // 1 year from now
				payload := unifiedPaymentModel.GeneratePaymentTokenRequest{
					PaymentID: paymentID,
					ExpiryAt:  futureTime.Format(time.RFC3339),
				}
				jsonPayload, err := json.Marshal(payload)
				assert.NoError(t, err)
				return jsonPayload
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var apiResponse struct {
					Code    string                                           `json:"code"`
					Message string                                           `json:"message"`
					Data    unifiedPaymentModel.GeneratePaymentTokenResponse `json:"data"`
				}
				err := json.NewDecoder(recorder.Body).Decode(&apiResponse)
				assert.NoError(t, err)
				assert.Equal(t, generatedToken, apiResponse.Data.Token)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := validator.New()
			paymentSvc := mocks.NewIPaymentService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(paymentSvc)

			// Statsd Monitoring
			monitor, err := pkgMonitor.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			// Create the controller instance
			merchantSvc := mocks.NewIMerchantService(t)
			creditcardSvc := mocks.NewICreditCardService(t)
			orchestratorSvc := mocks.NewIOrchestratorService(t)
			customerSvc := mocks.NewICustomerService(t)
			paymentMethodSvc := mocks.NewIPaymentMethodService(t)

			controller := New(config, mockValidator, mockLogger, monitor, Services{
				MerchantSvc:      merchantSvc,
				CreditcardSvc:    creditcardSvc,
				OrchestratorSvc:  orchestratorSvc,
				CustomerSvc:      customerSvc,
				PaymentMethodSvc: paymentMethodSvc,
				PaymentSvc:       paymentSvc,
			})

			baseUrl := "/api/v1/internal/cards/extend-payment-token"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GeneratePaymentToken)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)

			if tt.validateResp != nil {
				tt.validateResp(t, httpRecorder)
			}

			paymentSvc.AssertExpectations(t)
		})
	}
}
