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

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgMonitor "github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePayment(t *testing.T) {
	merchantID := uuid.New()
	validMerchant := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: merchantID.String(),
	}

	validPayload := creditcardModel.CreateCardPaymentRequest{
		ReferenceID:          "reference-id",
		BankMerchantID:       "bank-merchant-id",
		Amount:               decimal.NewFromFloat(10000.0),
		Currency:             "IDR",
		AuthenticationMethod: constant.CreditCardMethodChallenge,
		RedirectUrl: creditcardModel.CreditcardRedirectUrlRequest{
			SuccessUrl: "http://example.com/success",
			FailedUrl:  "http://example.com/failed",
		},
	}

	invalidAuthenticationMethodPayload := creditcardModel.CreateCardPaymentRequest{
		ReferenceID:          "reference-id",
		BankMerchantID:       "bank-merchant-id",
		Amount:               decimal.NewFromFloat(10000.0),
		Currency:             "IDR",
		AuthenticationMethod: "some-method",
	}

	config := &config.Config{
		CreditcardConfig: config.CreditcardConfig{
			WebviewURL: "https://example.com",
		},
	}

	testCases := []struct {
		name           string
		mockSetup      func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		merchantClaim  *merchantModel.MerchantAuthTokenClaims
		setHeaders     func(req *http.Request)
	}{
		{
			name: "SUCCESS: Create Payment",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {
				creditcardSvc.
					On("CreatePayment",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("card.CreateCardPaymentRequest")).
					Return(&creditcardModel.CreateCardPaymentResponse{
						ReferenceID: "reference-id",
						PaymentURL:  "https://example.com/pay/123",
					}, nil)
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "SUCCESS: Create Payment in behalf of submerchant",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {
				creditcardSvc.
					On("CreatePayment",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("card.CreateCardPaymentRequest")).
					Return(&creditcardModel.CreateCardPaymentResponse{
						ReferenceID: "reference-id",
						PaymentURL:  "https://example.com/pay/123",
					}, nil)
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "ERROR: Invalid Authentication Method",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(invalidAuthenticationMethodPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name:      "ERROR: Invalid JSON",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {},
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name:      "ERROR: Validation Failure",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := creditcardModel.CreateCardPaymentRequest{
					ReferenceID: "",
					Amount:      decimal.NewFromFloat(0),
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name:      "ERROR: Amount Less Than Minimum",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := creditcardModel.CreateCardPaymentRequest{
					ReferenceID: "reference-id",
					Amount:      decimal.NewFromFloat(9.99),
					Currency:    "IDR",
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name:      "ERROR: Invalid Authentication Method",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {},
			setupBody: func(t *testing.T) []byte {
				invalidPayload := creditcardModel.CreateCardPaymentRequest{
					ReferenceID:          "reference-id",
					Amount:               decimal.NewFromFloat(100.0),
					Currency:             "IDR",
					AuthenticationMethod: "invalid-method",
				}
				payload, err := json.Marshal(invalidPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name:      "ERROR: Unauthorized",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusUnauthorized,
			merchantClaim:  nil,
		},
		{
			name: "ERROR: Service Error",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService) {
				creditcardSvc.
					On("CreatePayment",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("card.CreateCardPaymentRequest")).
					Return(nil, errors.New("service error"))
			},
			setupBody: func(t *testing.T) []byte {
				payload, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payload
			},
			expectedStatus: http.StatusInternalServerError,
			merchantClaim:  validMerchant,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := validator.New()
			creditcardSvc := mocks.NewICreditCardService(t)
			orchestratorSvc := mocks.NewIOrchestratorService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(creditcardSvc, orchestratorSvc)

			// Statsd Monitoring
			monitor, err := pkgMonitor.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			// Create the controller instance
			merchantSvc := mocks.NewIMerchantService(t)
			customerSvc := mocks.NewICustomerService(t)
			controller := New(config, mockValidator, mockLogger, monitor, Services{
				MerchantSvc:     merchantSvc,
				CreditcardSvc:   creditcardSvc,
				OrchestratorSvc: orchestratorSvc,
				CustomerSvc:     customerSvc,
			})

			baseUrl := "/api/internal/v1/payments/create"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tt.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tt.merchantClaim))
			}

			if tt.setHeaders != nil {
				tt.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.CreatePayment)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			creditcardSvc.AssertExpectations(t)
		})
	}
}
