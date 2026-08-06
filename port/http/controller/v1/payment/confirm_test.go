package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmPayment(t *testing.T) {
	var (
		mockPaymentService        = mockService.NewIPaymentService(t)
		mockUnifiedPaymentService = mockService.NewIUnifiedPaymentService(t)
		controller                = PaymentController{
			config: &config.Config{
				UnifiedPaymentConfig: config.UnifiedPaymentConfig{
					VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
						MinAmount: func() *float64 { v := 10000.0; return &v }(),
						MaxAmount: func() *float64 { v := 10000000.0; return &v }(),
					},
					QrConfig: &config.UnifiedPaymentQrConfig{
						MinAmount: func() *float64 { v := 1000.0; return &v }(),
						MaxAmount: func() *float64 { v := 5000000.0; return &v }(),
					},
				},
			},
			validate:              validator.New(),
			monitor:               &monitoring.Monitor{},
			paymentService:        mockPaymentService,
			unifiedPaymentService: mockUnifiedPaymentService,
			logger:                func() logger.ILogger { logger, _ := logger.NewZapLogger(logger.Config{}); return logger }(),
		}
	)

	testCases := []struct {
		name           string
		setupContext   func() context.Context
		requestBody    string
		callMock       func()
		expectedStatus int
		validateResp   func(*testing.T, []byte)
	}{
		{
			name: "SUCCESS: Confirm payment with VA",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxPaymentID, "test-payment-123")
			},
			requestBody: `{
				"paymentMethod": {
					"type": "VIRTUAL_ACCOUNT"
				},
				"paymentMethodOptions": {
					"virtualAccount": {
						"channel": "BRI"
					}
				}
			}`,
			callMock: func() {
				// Mock GetPaymentDetailForPaymentUI first (to get merchant ID)
				mockPaymentService.On("GetPaymentDetailForPaymentUI", mock.Anything, "test-payment-123").
					Return(&paymentModel.PaymentDetailForPaymentUIResponse{
						UUID:       "test-payment-123",
						MerchantID: "merchant-456",
					}, nil).Once()

				// Mock GetSessionDetail (unified payment service)
				mockUnifiedPaymentService.On("GetSessionDetail", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.GetUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "test-payment-123" && req.MerchantID == "merchant-456"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					ID:     "test-payment-123",
					Status: "PENDING",
					Amount: unifiedPaymentModel.Amount{
						Value:    50000,
						Currency: "IDR",
					},
					PaymentType: constant.UnifiedPaymentTypeSingle,
				}, nil).Once()

				mockUnifiedPaymentService.On("ConfirmSession", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "test-payment-123" &&
						req.MerchantID == "merchant-456" &&
						req.PaymentMethod.Type == constant.UnifiedPaymentMethodVA
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					ID:                "test-payment-123",
					ClientReferenceID: "ref-789",
					Status:            "ACTIVE",
					Amount: unifiedPaymentModel.Amount{
						Value:    50000,
						Currency: "IDR",
					},
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "00", response["code"])
				assert.NotNil(t, response["data"])
			},
		},
		{
			name: "ERROR: Missing payment ID in context",
			setupContext: func() context.Context {
				return context.Background() // No payment ID
			},
			requestBody: `{
				"paymentMethod": {
					"type": "VIRTUAL_ACCOUNT"
				}
			}`,
			callMock:       func() {}, // No service calls expected
			expectedStatus: http.StatusUnauthorized,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "41", response["code"])
			},
		},
		{
			name: "ERROR: Payment not found",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxPaymentID, "non-existent")
			},
			requestBody: `{
				"paymentMethod": {
					"type": "VIRTUAL_ACCOUNT"
				}
			}`,
			callMock: func() {
				mockPaymentService.On("GetPaymentDetailForPaymentUI", mock.Anything, "non-existent").
					Return(nil, pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "44", response["code"])
			},
		},
		{
			name: "ERROR: Invalid JSON payload",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxPaymentID, "test-payment")
			},
			requestBody:    `invalid-json`,
			callMock:       func() {}, // No service calls expected
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "40", response["code"])
			},
		},
		{
			name: "ERROR: Validation error - missing VA options",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxPaymentID, "test-validation")
			},
			requestBody: `{
				"paymentMethod": {
					"type": "VIRTUAL_ACCOUNT"
				}
			}`,
			callMock: func() {
				mockPaymentService.On("GetPaymentDetailForPaymentUI", mock.Anything, "test-validation").
					Return(&paymentModel.PaymentDetailForPaymentUIResponse{
						UUID:       "test-validation",
						MerchantID: "test-merchant",
					}, nil).Once()

				mockUnifiedPaymentService.On("GetSessionDetail", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.GetUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "test-validation" && req.MerchantID == "test-merchant"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					ID:     "test-validation",
					Status: "PENDING",
					Amount: unifiedPaymentModel.Amount{
						Value:    25000,
						Currency: "IDR",
					},
					PaymentType:    constant.UnifiedPaymentTypeSingle,
					ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
				}, nil).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "45", response["code"])
			},
		},
		{
			name: "ERROR: Unified payment service error",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxPaymentID, "test-service-error")
			},
			requestBody: `{
				"paymentMethod": {
					"type": "VIRTUAL_ACCOUNT"
				},
				"paymentMethodOptions": {
					"virtualAccount": {
						"channel": "BCA"
					}
				}
			}`,
			callMock: func() {
				mockPaymentService.On("GetPaymentDetailForPaymentUI", mock.Anything, "test-service-error").
					Return(&paymentModel.PaymentDetailForPaymentUIResponse{
						UUID:       "test-service-error",
						MerchantID: "test-merchant",
					}, nil).Once()

				mockUnifiedPaymentService.On("GetSessionDetail", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.GetUnifiedPaymentSessionRequest) bool {
					return req.PaymentSessionID == "test-service-error" && req.MerchantID == "test-merchant"
				})).Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
					ID:     "test-service-error",
					Status: "PENDING",
					Amount: unifiedPaymentModel.Amount{
						Value:    30000,
						Currency: "IDR",
					},
					PaymentType:    constant.UnifiedPaymentTypeSingle,
					ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
				}, nil).Once()

				mockUnifiedPaymentService.On("ConfirmSession", mock.Anything, mock.Anything).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "general_error", response["code"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			req := httptest.NewRequest(http.MethodPost, "/payments/confirm", bytes.NewBufferString(tc.requestBody))
			req = req.WithContext(tc.setupContext())
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			controller.ConfirmPayment(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.validateResp != nil {
				tc.validateResp(t, w.Body.Bytes())
			}

			mockPaymentService.AssertExpectations(t)
			mockUnifiedPaymentService.AssertExpectations(t)
		})
	}
}

func TestValidateConfirmPayload(t *testing.T) {
	controller := &PaymentController{
		config: &config.Config{
			UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
					MinAmount: func() *float64 { v := 10000.0; return &v }(),
					MaxAmount: func() *float64 { v := 10000000.0; return &v }(),
				},
				QrConfig: &config.UnifiedPaymentQrConfig{
					MinAmount: func() *float64 { v := 1000.0; return &v }(),
					MaxAmount: func() *float64 { v := 5000000.0; return &v }(),
				},
			},
		},
	}

	testCases := []struct {
		name    string
		payload *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Valid VA payload",
			payload: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "BRI",
					},
				},
				Amount: unifiedPaymentModel.Amount{
					Value:    50000,
					Currency: "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Missing VA options",
			payload: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				Amount: unifiedPaymentModel.Amount{
					Value:    30000,
					Currency: "IDR",
				},
			},
			wantErr: true,
		},
		{
			name: "ERROR: Amount below minimum",
			payload: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "BCA",
					},
				},
				Amount: unifiedPaymentModel.Amount{
					Value:    5000, // Below minimum
					Currency: "IDR",
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := controller.validateConfirmPayload(tc.payload)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
