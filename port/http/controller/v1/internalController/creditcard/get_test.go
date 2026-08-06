package creditcard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
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

func TestGetPaymentById(t *testing.T) {
	uuid := uuid.New()
	referenceId := "reference-id"
	now := time.Now()
	expiredAt := now.Add(10 * time.Minute)
	validMerchant := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: "merchant-id",
	}

	creditCard := &paymentModel.Payment{
		UUID:            uuid.String(),
		ReferenceID:     &referenceId,
		CustomerID:      "customer-id",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		Amount:          decimal.NewFromFloat(10000),
		Currency:        "IDR",
		PaymentURL:      "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/hOIuKiu-6NxhiFnJPMDWIke9qq0YsbpERh4Atnn-AEY=",
		Status:          "PENDING",
		ExpiredAt:       &expiredAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	config := &config.Config{
		CreditcardConfig: config.CreditcardConfig{
			WebviewURL: "https://example.com",
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(
			creditcardSvc *mocks.ICreditCardService,
			orchestratorSvc *mocks.IOrchestratorService,
			merchantSvc *mocks.IMerchantService,
			customerSvc *mocks.ICustomerService,
			paymentMethodSvc *mocks.IPaymentMethodService)
		expectedStatus int
		merchantClaim  *merchantModel.MerchantAuthTokenClaims
		uuid           string
		isNetworkToken string
	}{
		{
			name: "SUCCESS: Get Payment By Id with all data",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://test.example.com",
										AcquirerMerchantID: "test-acquirer-id",
										Priority:           1,
										IsActive:           true,
									},
								},
							},
						},
					},
				}, nil)
				bypassFlag := true
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(&merchantModel.GetFDSConfigResponse{
						FDSConfig: merchantModel.FDSConfig{
							BypassExternalPaymentCheck: &bypassFlag,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "SUCCESS: Get Payment By Id - customer fetch fails but continues",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(nil, errors.New("customer not found"))
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{}, nil)
				bypassFlag := true
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(&merchantModel.GetFDSConfigResponse{
						FDSConfig: merchantModel.FDSConfig{
							BypassExternalPaymentCheck: &bypassFlag,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "SUCCESS: Get Payment By Id with sub-merchant parent config",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				metadataMap := map[string]any{
					"onBehalf": map[string]any{
						"parentMerchantId": "parent-merchant-id",
					},
				}
				creditCardWithOnBehalf := &paymentModel.Payment{
					UUID:            uuid.String(),
					ReferenceID:     &referenceId,
					CustomerID:      "customer-id",
					MerchantID:      "merchant-id",
					PaymentMethodID: "payment-method-id",
					Amount:          decimal.NewFromFloat(10000),
					Currency:        "IDR",
					PaymentURL:      "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/test",
					Status:          "PENDING",
					ExpiredAt:       &expiredAt,
					CreatedAt:       now,
					UpdatedAt:       now,
					Metadata:        &metadataMap,
				}
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCardWithOnBehalf, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				// Child merchant payment method
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						"payment-method-id",
						"merchant-id",
					).Return(&paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://child.example.com",
										AcquirerMerchantID: "child-acquirer-id",
										Priority:           1,
										IsActive:           true,
									},
								},
							},
						},
					},
				}, nil)
				// Parent merchant payment method
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						"payment-method-id",
						"parent-merchant-id",
					).Return(&paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://parent.example.com",
										AcquirerMerchantID: "parent-acquirer-id",
										Priority:           1,
										IsActive:           true,
									},
								},
							},
						},
					},
				}, nil)
				bypassFlag := false
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(&merchantModel.GetFDSConfigResponse{
						FDSConfig: merchantModel.FDSConfig{
							BypassExternalPaymentCheck: &bypassFlag,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "SUCCESS: Get Payment By Id - FDS config returns nil",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{}, nil)
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name:           "SUCCESS: Get Payment By Id with network token query param",
			uuid:           uuid.String(),
			isNetworkToken: "true",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://test.example.com",
										AcquirerMerchantID: "default-acquirer-mid",
										MerchantIDTag:      "default-tag",
										Priority:           1,
										IsActive:           true,
										NetworkToken:       &paymentMethodModel.CardNetworkTokenPartnerConfigObj{Type: "DEFAULT"},
									},
								},
							},
						},
					},
				}, nil)
				bypassFlag := true
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(&merchantModel.GetFDSConfigResponse{
						FDSConfig: merchantModel.FDSConfig{
							BypassExternalPaymentCheck: &bypassFlag,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name:           "SUCCESS: Get Payment By Id with network token and COF initiator",
			uuid:           uuid.String(),
			isNetworkToken: "true",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				// Payment with cardOnFile metadata containing MERCHANT initiator
				metadataMap := map[string]any{
					"paymentMethodOptions": map[string]any{
						"card": map[string]any{
							"cardOnFile": map[string]any{
								"initiator": "MERCHANT",
							},
						},
					},
				}
				creditCardWithCOF := &paymentModel.Payment{
					UUID:            uuid.String(),
					ReferenceID:     &referenceId,
					CustomerID:      "customer-id",
					MerchantID:      "merchant-id",
					PaymentMethodID: "payment-method-id",
					Amount:          decimal.NewFromFloat(10000),
					Currency:        "IDR",
					PaymentURL:      "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/test",
					Status:          "PENDING",
					ExpiredAt:       &expiredAt,
					CreatedAt:       now,
					UpdatedAt:       now,
					Metadata:        &metadataMap,
				}
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCardWithCOF, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://test.example.com",
										AcquirerMerchantID: "default-mid",
										Priority:           1,
										IsActive:           true,
										NetworkToken:       &paymentMethodModel.CardNetworkTokenPartnerConfigObj{Type: "DEFAULT"},
									},
									{
										PartnerProcessor:   "MPGS",
										PartnerBaseURL:     "https://test.example.com",
										AcquirerMerchantID: "cof-merchant-mid",
										MerchantIDTag:      "cof-merchant-tag",
										Priority:           2,
										IsActive:           true,
										NetworkToken:       &paymentMethodModel.CardNetworkTokenPartnerConfigObj{Type: "COF", COFInitiator: "MERCHANT"},
									},
								},
							},
						},
					},
				}, nil)
				bypassFlag := true
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(&merchantModel.GetFDSConfigResponse{
						FDSConfig: merchantModel.FDSConfig{
							BypassExternalPaymentCheck: &bypassFlag,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "SUCCESS: Get Payment By Id for Virtual Terminal Transaction",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.On(
					"GetPaymentById", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.Payment{
					UUID:            uuid.String(),
					ReferenceID:     &referenceId,
					CustomerID:      "customer-id",
					MerchantID:      "merchant-id",
					PaymentMethodID: "payment-method-id",
					Amount:          decimal.NewFromFloat(10000),
					Currency:        "IDR",
					PaymentURL:      "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/hOIuKiu-6NxhiFnJPMDWIke9qq0YsbpERh4Atnn-AEY=",
					Status:          "PENDING",
					ExpiredAt:       &expiredAt,
					Metadata: &map[string]any{
						"virtualTerminal": map[string]any{
							"travelAgentCode":    "BOOKING",                             // NOSONAR
							"allowedCardTypes":   []string{"CREDIT", "DEBIT"},           // NOSONAR
							"allowedPrincipal":   []string{"VISA", "MASTERCARD", "JCB"}, // NOSONAR
							"allowedBinNumbers":  []string{"444810"},                    // NOSONAR
							"acquirerMerchantId": "TEST001101217287",                    // NOSONAR
						},
					},
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				orchestratorSvc.On(
					"FindByReference", mock.Anything, uuid.String(), "PAYMENT",
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid}, nil)
				customerSvc.On(
					"GetCustomerById", mock.Anything, mock.Anything, mock.Anything,
				).Return(nil, nil)
				paymentMethodSvc.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{}, nil)
				merchantSvc.On(
					"GetFDSConfig", mock.Anything, mock.Anything,
				).Return(&merchantModel.GetFDSConfigResponse{
					FDSConfig: merchantModel.FDSConfig{
						BypassExternalPaymentCheck: new(true),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchant,
		},
		{
			name: "ERROR: Unauthorized",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
			},
			expectedStatus: http.StatusUnauthorized,
			merchantClaim:  nil,
			uuid:           "some-uuid",
		},
		{
			name: "ERROR: Missing UUID",
			uuid: "",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
			},
			expectedStatus: http.StatusBadRequest,
			merchantClaim:  validMerchant,
		},
		{
			name: "ERROR: CreditCard Service Error",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			merchantClaim:  validMerchant,
			uuid:           "some-uuid",
		},
		{
			name: "ERROR: Orchestrator Service Error",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(nil, errors.New("orchestrator error"))
			},
			expectedStatus: http.StatusInternalServerError,
			merchantClaim:  validMerchant,
		},
		{
			name: "ERROR: PaymentMethod Service Error",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(nil, errors.New("payment method error"))
			},
			expectedStatus: http.StatusInternalServerError,
			merchantClaim:  validMerchant,
		},
		{
			name: "ERROR: FDS Config fetch error",
			uuid: uuid.String(),
			mockSetup: func(creditcardSvc *mocks.ICreditCardService, orchestratorSvc *mocks.IOrchestratorService, merchantSvc *mocks.IMerchantService, customerSvc *mocks.ICustomerService, paymentMethodSvc *mocks.IPaymentMethodService) {
				creditcardSvc.
					On("GetPaymentById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(creditCard, nil)
				orchestratorSvc.
					On("FindByReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						uuid.String(),
						"PAYMENT").
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID: uuid,
					}, nil)
				customerResp := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid",
					MerchantID:  "merchant-id",
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}
				customerSvc.
					On("GetCustomerById",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).
					Return(customerResp, nil)
				paymentMethodSvc.
					On("FindPaymentMethodByIdAndMerchant",
						constant.ValueCtxMockType(),
						constant.StringMockType(),
						constant.StringMockType(),
					).Return(&paymentModel.PaymentMethodWithPivot{}, nil)
				merchantSvc.
					On("GetFDSConfig",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string")).
					Return(nil, errors.New("fds config error"))
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
			merchantSvc := mocks.NewIMerchantService(t)
			customerSvc := mocks.NewICustomerService(t)
			paymentMethodSvc := mocks.NewIPaymentMethodService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(creditcardSvc, orchestratorSvc, merchantSvc, customerSvc, paymentMethodSvc)

			// Statsd Monitoring
			monitor, err := pkgMonitor.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			// Create the controller instance
			controller := New(config, mockValidator, mockLogger, monitor, Services{
				MerchantSvc:      merchantSvc,
				CreditcardSvc:    creditcardSvc,
				OrchestratorSvc:  orchestratorSvc,
				CustomerSvc:      customerSvc,
				PaymentMethodSvc: paymentMethodSvc,
			})

			baseUrl := "/api/internal/v1/payments/{uuid}"
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("uuid", tt.uuid)
			reqURL := baseUrl
			if tt.isNetworkToken != "" {
				reqURL = reqURL + "?isNetworkToken=" + tt.isNetworkToken
			}
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tt.merchantClaim))
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetPaymentById)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			creditcardSvc.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}

func TestGetStoredCardByCustomerID(t *testing.T) {
	config := &config.Config{
		CreditcardConfig: config.CreditcardConfig{
			WebviewURL: "https://example.com",
		},
	}

	testCases := []struct {
		name           string
		mockSetup      func(creditcardSvc *mocks.ICreditCardService)
		expectedStatus int
		merchantId     string
		customerId     string
	}{
		{
			name: "SUCCESS: Get stored cards by customer ID",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService) {
				mockCards := []*unifiedPaymentModel.CustomerPaymentMethodResponse{
					{
						Token:          "card-token-123",
						PaymentMethod:  constant.UnifiedPaymentMethodCard,
						PaymentChannel: "VISA",
						Status:         constant.StoredPaymentMethodStatusActive,
						CreatedAt:      time.Now(),
						Card: &unifiedPaymentModel.CustomerPaymentMethodCardResponse{
							Fingerprint: "fp-123",
							Network:     "VISA",
							Last4:       "1234",
							ExpMonth:    "12",
							ExpYear:     "2025",
						},
					},
					{
						Token:          "card-token-456",
						PaymentMethod:  constant.UnifiedPaymentMethodCard,
						PaymentChannel: "MASTERCARD",
						Status:         constant.StoredPaymentMethodStatusActive,
						CreatedAt:      time.Now(),
						Card: &unifiedPaymentModel.CustomerPaymentMethodCardResponse{
							Fingerprint: "fp-456",
							Network:     "MASTERCARD",
							Last4:       "5678",
							ExpMonth:    "06",
							ExpYear:     "2026",
						},
					},
				}
				creditcardSvc.
					On("GetStoredCardByCustomerID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						"merchant-123",
						"customer-456").
					Return(mockCards, nil)
			},
			expectedStatus: http.StatusOK,
			merchantId:     "merchant-123",
			customerId:     "customer-456",
		},
		{
			name: "SUCCESS: Get stored cards - empty result",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService) {
				emptyCards := []*unifiedPaymentModel.CustomerPaymentMethodResponse{}
				creditcardSvc.
					On("GetStoredCardByCustomerID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						"merchant-123",
						"customer-789").
					Return(emptyCards, nil)
			},
			expectedStatus: http.StatusOK,
			merchantId:     "merchant-123",
			customerId:     "customer-789",
		},
		{
			name:           "ERROR: Missing merchant ID",
			mockSetup:      func(creditcardSvc *mocks.ICreditCardService) {},
			expectedStatus: http.StatusBadRequest,
			merchantId:     "",
			customerId:     "customer-456",
		},
		{
			name:           "ERROR: Missing customer ID",
			mockSetup:      func(creditcardSvc *mocks.ICreditCardService) {},
			expectedStatus: http.StatusBadRequest,
			merchantId:     "merchant-123",
			customerId:     "",
		},
		{
			name: "ERROR: Service error",
			mockSetup: func(creditcardSvc *mocks.ICreditCardService) {
				creditcardSvc.
					On("GetStoredCardByCustomerID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						"merchant-123",
						"customer-error").
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			merchantId:     "merchant-123",
			customerId:     "customer-error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := validator.New()
			creditcardSvc := mocks.NewICreditCardService(t)
			orchestratorSvc := mocks.NewIOrchestratorService(t)
			merchantSvc := mocks.NewIMerchantService(t)
			customerSvc := mocks.NewICustomerService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(creditcardSvc)

			// Statsd Monitoring
			monitor, err := pkgMonitor.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			// Create the controller instance
			controller := New(config, mockValidator, mockLogger, monitor, Services{
				MerchantSvc:     merchantSvc,
				CreditcardSvc:   creditcardSvc,
				OrchestratorSvc: orchestratorSvc,
				CustomerSvc:     customerSvc,
			})

			baseUrl := "/api/internal/v1/merchants/{merchantId}/customers/{customerId}/cards"
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("merchantId", tt.merchantId)
			chiRouterCtx.URLParams.Add("customerId", tt.customerId)
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetStoredCardByCustomerID)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, httpRecorder.Code)
			creditcardSvc.AssertExpectations(t)
		})
	}
}
