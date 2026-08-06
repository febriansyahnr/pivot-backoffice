package fdsservice

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestCheckTransaction(t *testing.T) {
	// Create default values for reuse
	transactionID := "test-transaction-id"
	merchantID := uuid.New()
	referenceID := "test-reference-id"
	trxUUID := uuid.New()

	// Mock payment metadata
	mockMetadata := map[string]interface{}{
		"cardData": map[string]interface{}{
			"Fingerprint":     "test-fingerprint",
			"CardFingerprint": "test-card-fingerprint",
			"CardUuid":        "test-card-uuid",
			"First8Digit":     "12345678",
			"Last4Digit":      "1234",
			"CardBrand":       "VISA",
		},
		"authenticationData": map[string]interface{}{
			"EciCode": "05",
			"XID":     "test-xid",
		},
		"authorizationData": map[string]interface{}{
			"AcquirerTransactionID": "test-acquirer-id",
			"CvvResult":             "M",
			"ApprovalCode":          "123456",
			"AcquirerResponseCode":  "00",
		},
	}

	raw := `{
    "expiredAt": "2025-05-30T15:04:05Z",
    "feeDetail": {
        "type": "PAYMENT",
        "amount": 2000,
        "method": "CREDIT_CARD",
        "taxType": "NON_PKP",
        "taxAmount": 0,
        "amountType": "AMOUNT_PERCENTAGE",
        "percentage": 2.5,
        "finalAmount": 4500,
        "deductionType": "DIRECT",
        "referenceType": "",
        "taxPercentage": 0
    },
    "chargeStatus": "SUCCESS",
    "methodDetail": {
        "card": {
            "last4": "0010",
            "first6": "444000",
            "first8": "44400000",
            "expYear": 0,
            "expMonth": 0,
            "fingerprint": "0077187d-f69d-4b2c-a9f9-99aeb6919dda",
            "cardUuid": "test-card-uuid-from-raw-json",
			"cardBrand": "VISA",
			"countryCode": "ID",
            "binInformations": {
                "type": "DEBIT",
                "brand": "VISA",
                "country": "ID",
                "issuingBank": "BRI_S2I"
            },
            "authorizationResult": {
                "stan": "255002",
                "avsResult": "",
                "cvvResult": "",
                "authorizedAmount": {
                    "value": 100000,
                    "currency": "IDR"
                },
                "acquirerReferenceNumber": "123456789012345",
                "issuerAuthorizationCode": "00",
                "retrievalReferenceNumber": "TRXCC91a3b126dc1b17461525831"
            },
            "authenticationResult": {
                "eciCode": "05",
                "threeDsMethod": "",
                "threeDsResult": "AUTHENTICATION_SUCCESSFUL",
                "threeDsVersion": "2.2.0"
            }
        }
    },
    "reconReferenceNo": "PAYCC5fd2849b40191746152709",
    "settlementDetail": {
        "type": "T+5",
        "dayType": "ANYDAY",
        "endCutOffTime": "",
        "executionTime": "",
        "startCutOffTime": ""
    },
    "processorTransactionId": ""
}`

	additionalInfo := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText([]byte(raw)),
	}

	customerRepo := mockRepositories.NewICustomerRepository(t)

	testCases := []struct {
		desc         string
		wantErr      bool
		expectedResp *fdscommon.CheckTransactionResponse
		request      *fdscommon.CheckTransactionRequest
		setupMock    func(
			accountTrxRepo *mockRepositories.IAccountTransactionRepository,
			paymentRepo *mockRepositories.IPaymentRepository,
			merchantRepo *mockRepositories.IMerchantRepository,
			fraudRulesRepo *mockRepositories.IFraudRulesRepository,
			ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
			fraudNetMock *mockRepositories.IFdsProcessorRepository,
			rabbitMq *mockRabbitMq.RabbitMQExt,
		)
	}{
		{
			desc: "success case with low risk score",
			request: &fdscommon.CheckTransactionRequest{
				BillingInformation: fdscommon.BillingInformation{
					Email:     "example@email.com", // NOSONAR
					GivenName: "John",              // NOSONAR
					Surname:   "Doe",               // NOSONAR
				},
			},
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(r *fdscommon.CheckRequest) bool {
					return r.Customer.Email != nil &&
						r.Customer.FirstName != nil &&
						r.Customer.LastName != nil &&
						*r.Customer.Email == "example@email.com" &&
						*r.Customer.FirstName == "John" &&
						*r.Customer.LastName == "Doe"
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case with high risk score (rejected)",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REJECTED,
				Score:  decimal.NewFromInt(60),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules with higher weight
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.6),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response with high risk
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "HIGH",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)

				// Mock RabbitMQ publish for Slack notification (since status will be REJECTED)
				rabbitMq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("[]uint8")).Return(nil)
			},
		},
		{
			desc:    "error case - transaction not found",
			wantErr: true,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction not found
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:    "error case - payment not found",
			wantErr: true,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment not found
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:    "error case - merchant not found",
			wantErr: true,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant not found
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:    "error case - fraud rules not found",
			wantErr: true,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock empty fraud rules
				emptyRules := []*fraudrulesmodel.FraudRules{}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(emptyRules, 0, nil)
			},
		},
		{
			desc:    "error case - fraud processor check fails",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor check failure
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(nil, errors.New(response.HttpErrInternal, constant.ErrFraudRulesNotFound))

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "error case - rule evaluation creation fails",
			wantErr: false, // No overall error since we continue processing
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation creation failure
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(errors.New(response.HttpErrInternal, constant.ErrFraudRulesNotFound))

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - multiple rules with mixed results",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(45),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock multiple fraud rules with different weights
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule 1",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
					{
						UUID:          "rule-2",
						RuleName:      "Test Rule 2",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.15),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response for first rule
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// This is a simple way to differentiate the calls, you might need a more robust approach
					return true // For the first call
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100, // 100 * 0.3 = 30
						RiskGroup: "LOW",
					},
				}, nil).Once()

				// Mock processor response for second rule
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// For the second call
					return true
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100, // 100 * 0.15 = 15
						RiskGroup: "MEDIUM",
					},
				}, nil).Once()

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil).Twice()

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - processor returns unsuccessful response",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response - unsuccessful
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: false,
					Code:    util.ValueToPtr("error"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"error": "validation failed",
					},
				}, nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "error case - failed to get creditcard metadata",
			wantErr: true,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_NOT_EVALUATED,
				Score:  decimal.NewFromInt(0),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction with invalid additional info that causes GetCreditcardMetadataFromAdditionalInfo to fail
				invalidAdditionalInfo := types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"invalid": "json structure for creditcard"}`)),
				}

				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: invalidAdditionalInfo,
				}, nil)
			},
		},
		{
			desc:    "success case - rule with invalid provider should be skipped",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules - one with invalid provider, one with valid provider
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule Invalid Provider",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							Valid: false, // Invalid provider - should be skipped
						},
						Weight: decimal.NewFromFloat(0.2),
					},
					{
						UUID:          "rule-2",
						RuleName:      "Test Rule Valid Provider",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 2, nil)

				// Mock processor response for the valid rule only
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - rule with unknown provider should be skipped",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules - one with unknown provider, one with known provider
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule Unknown Provider",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: "UNKNOWN_PROVIDER", // Provider not in thirdPartyProcessor map
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.2),
					},
					{
						UUID:          "rule-2",
						RuleName:      "Test Rule Known Provider",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 2, nil)

				// Mock processor response for the known provider only
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - payment with customerID (uses customerID from request)",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment WITH CustomerID
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					CustomerID:  "customer-123", // This should be used instead of CardFingerprint
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response - verify that customer_id is "customer-123"
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// Check that Customer.ID uses the payment.CustomerID ("customer-123") not CardFingerprint
					return req.Customer.ID == "customer-123"
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
				customerRepo.On("GetCustomerById", mock.Anything, "customer-123", merchantID.String()).Return(nil, nil).Maybe()
			},
		},
		{
			desc:    "success case - payment without customerID (uses CardFingerprint)",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment WITHOUT CustomerID (empty string)
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					CustomerID:  "", // Empty - should use CardFingerprint
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response - verify that customer_id uses CardFingerprint
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// Check that Customer.ID uses CardFingerprint since payment.CustomerID is empty
					return req.Customer.ID == "0077187d-f69d-4b2c-a9f9-99aeb6919dda"
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - with device information in request",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			request: &fdscommon.CheckTransactionRequest{
				Device: &fdscommon.DeviceCheck{
					IPAddress:           util.ValueToPtr("103.87.202.115"),
					UserAgent:           util.ValueToPtr("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"),
					FingerprintID:       util.ValueToPtr("4f83661117c8611d27055867ad292c25"),
					SessionID:           util.ValueToPtr("session_mesmi1oi20c5kgqvuo"),
					ClientID:            util.ValueToPtr("client_meuqxr9867jkfmuiwlr"),
					Resolution:          util.ValueToPtr("1680×1050"),
					TimeZone:            util.ValueToPtr("Asia/Jakarta"),
					Language:            util.ValueToPtr("en-US"),
					IsProxy:             util.ValueToPtr(true),
					ColorDepth:          util.ValueToPtr(30),
					FontSmoothing:       util.ValueToPtr(true),
					JavaSupport:         util.ValueToPtr(false),
					TouchSupport:        util.ValueToPtr(false),
					CookieSupport:       util.ValueToPtr(true),
					CanvasFingerprintID: util.ValueToPtr("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
					CanvasHeight:        util.ValueToPtr(50),
					CanvasWidth:         util.ValueToPtr(200),
					ScreenHeight:        util.ValueToPtr(1050),
					ScreenWidth:         util.ValueToPtr(1680),
					NumFonts:            util.ValueToPtr(4),
					FontsHash:           util.ValueToPtr("2c5269514ca62d898365fd6bb2d3a3f7432d7a6bbdd7233e5edd8b21b88ff3a9"),
					NumPlugins:          util.ValueToPtr(5),
					PluginHash:          util.ValueToPtr("6b13d24b814d5e1f3d5d72822d82254476d2054c431d02122b44a2f0932e20b0"),
					PluginsHash:         util.ValueToPtr("6b13d24b814d5e1f3d5d72822d82254476d2054c431d02122b44a2f0932e20b0"),
					MIMETypesHash:       util.ValueToPtr("d28f781d6ad647e86230f940ad30d7e9382f1ce58e7fb545c26344fe99684774"),
					NumMIMETypes:        util.ValueToPtr("2"),
					ProxyType:           util.ValueToPtr("proxy"),
					IsTor:               util.ValueToPtr(false),
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				paymentRepo.On("UpdatePaymentMetadataById", mock.Anything, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
					FingerprintID: "4f83661117c8611d27055867ad292c25",
				}).Return(nil)

				// Mock processor response - verify device information is passed through
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// Verify device information is properly set
					return req.Device.IPAddress != nil && *req.Device.IPAddress == "103.87.202.115" &&
						req.Device.UserAgent != nil && *req.Device.UserAgent == "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36" &&
						req.Device.FingerprintID != nil && *req.Device.FingerprintID == "4f83661117c8611d27055867ad292c25" &&
						req.Device.SessionID != nil && *req.Device.SessionID == "session_mesmi1oi20c5kgqvuo" &&
						req.Device.ClientID != nil && *req.Device.ClientID == "client_meuqxr9867jkfmuiwlr" &&
						req.Device.Resolution != nil && *req.Device.Resolution == "1680×1050" &&
						req.Device.TimeZone != nil && *req.Device.TimeZone == "Asia/Jakarta" &&
						req.Device.Language != nil && *req.Device.Language == "en-US" &&
						req.Device.IsProxy != nil && *req.Device.IsProxy == true &&
						req.Device.ColorDepth != nil && *req.Device.ColorDepth == 30 &&
						req.Device.FontSmoothing != nil && *req.Device.FontSmoothing == true &&
						req.Device.JavaSupport != nil && *req.Device.JavaSupport == false &&
						req.Device.TouchSupport != nil && *req.Device.TouchSupport == false &&
						req.Device.CookieSupport != nil && *req.Device.CookieSupport == true &&
						req.Device.CanvasFingerprintID != nil && *req.Device.CanvasFingerprintID == "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" &&
						req.Device.CanvasHeight != nil && *req.Device.CanvasHeight == 50 &&
						req.Device.CanvasWidth != nil && *req.Device.CanvasWidth == 200 &&
						req.Device.ScreenHeight != nil && *req.Device.ScreenHeight == 1050 &&
						req.Device.ScreenWidth != nil && *req.Device.ScreenWidth == 1680 &&
						req.Device.NumFonts != nil && *req.Device.NumFonts == 4 &&
						req.Device.FontsHash != nil && *req.Device.FontsHash == "2c5269514ca62d898365fd6bb2d3a3f7432d7a6bbdd7233e5edd8b21b88ff3a9" &&
						req.Device.NumPlugins != nil && *req.Device.NumPlugins == 5 &&
						req.Device.PluginHash != nil && *req.Device.PluginHash == "6b13d24b814d5e1f3d5d72822d82254476d2054c431d02122b44a2f0932e20b0" &&
						req.Device.PluginsHash != nil && *req.Device.PluginsHash == "6b13d24b814d5e1f3d5d72822d82254476d2054c431d02122b44a2f0932e20b0" &&
						req.Device.MIMETypesHash != nil && *req.Device.MIMETypesHash == "d28f781d6ad647e86230f940ad30d7e9382f1ce58e7fb545c26344fe99684774" &&
						req.Device.NumMIMETypes != nil && *req.Device.NumMIMETypes == "2" &&
						req.Device.ProxyType != nil && *req.Device.ProxyType == "proxy" &&
						req.Device.IsTor != nil && *req.Device.IsTor == false
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - fingerprint ID update succeeds",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			request: &fdscommon.CheckTransactionRequest{
				Device: &fdscommon.DeviceCheck{
					FingerprintID: util.ValueToPtr("test-fingerprint-id-12345"),
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock fingerprint ID update success (covers lines 91-93)
				paymentRepo.On("UpdatePaymentMetadataById", mock.Anything, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
					FingerprintID: "test-fingerprint-id-12345",
				}).Return(nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response
				fraudNetMock.On("Check", mock.Anything, mock.AnythingOfType("*fdscommon.CheckRequest")).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:         "error case - fingerprint ID update fails",
			wantErr:      true,
			expectedResp: nil, // Error case returns nil
			request: &fdscommon.CheckTransactionRequest{
				Device: &fdscommon.DeviceCheck{
					FingerprintID: util.ValueToPtr("test-fingerprint-id-12345"),
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock fingerprint ID update failure (covers lines 94-97)
				paymentRepo.On("UpdatePaymentMetadataById", mock.Anything, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
					FingerprintID: "test-fingerprint-id-12345",
				}).Return(errors.New(response.HttpErrInternal, constant.ErrPaymentNotFound))

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)
			},
		},
		{
			desc:         "error case - when failed to request update fingerprintID",
			wantErr:      true,
			expectedResp: nil, // Error case returns nil
			request: &fdscommon.CheckTransactionRequest{
				Device: &fdscommon.DeviceCheck{
					FingerprintID: util.ValueToPtr("test-fingerprint-id-1"),
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock fingerprint ID update failure (covers lines 94-97)
				paymentRepo.On("UpdatePaymentMetadataById", mock.Anything, payment.UUID, paymentModel.UpdatePaymentMetadataRequest{
					FingerprintID: "test-fingerprint-id-1",
				}).Return(errors.New(response.HttpErrInternal, constant.ErrPaymentNotFound))

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)
			},
		},
		{
			desc:    "success case - verify payment_id uses CardFingerprint",
			wantErr: false,
			expectedResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
				rabbitMq *mockRabbitMq.RabbitMQExt,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Channel:        "\"[ANY]\"",
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					CustomerID:  "",
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					Metadata:    &mockMetadata,
					CreatedAt:   time.Now(),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock merchant
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "Test",
					Address:   "Test Address",
					PostCode:  "12345",
					PICPhone:  "081234567890",
					PICEmail:  "test@example.com",
				}, nil)

				// Mock fraud rules
				rules := []*fraudrulesmodel.FraudRules{
					{
						UUID:          "rule-1",
						RuleName:      "Test Rule",
						ReferenceType: "\"[ANY]\"",
						Provider: sql.NullString{
							String: constant.PROVIDER_FRAUD_NET,
							Valid:  true,
						},
						Weight: decimal.NewFromFloat(0.3),
					},
				}
				fraudRulesRepo.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					ReferenceType: "\"[ANY]\"",
				}).Return(rules, 1, nil)

				// Mock processor response - verify payment_id uses CardFingerprint
				fraudNetMock.On("Check", mock.Anything, mock.MatchedBy(func(req *fdscommon.CheckRequest) bool {
					// Check that Payment.PaymentID uses CardFingerprint (new field)
					return req.Payment.PaymentID != nil && *req.Payment.PaymentID == "0077187d-f69d-4b2c-a9f9-99aeb6919dda"
				})).Return(&fdscommon.CheckResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{},
					Data: fdscommon.CheckData{
						RiskScore: 100,
						RiskGroup: "LOW",
					},
				}, nil)

				// Mock rule evaluation repo
				ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

				// Mock UpdateAdditionalInfoByID for saveFdsRiskAssessmentToLedger
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
			rabbitMq := mockRabbitMq.NewRabbitMQExt(t)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expected mocks
			tc.setupMock(accountTrxRepo, paymentRepo, merchantRepo, fraudRulesRepo, ruleEvalRepo, fraudNetMock, rabbitMq)

			// Create service with mocks
			processors := map[string]repository.IFdsProcessorRepository{
				constant.PROVIDER_FRAUD_NET: fraudNetMock,
			}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
					},
					SlackConfig: config.SlackConfig{
						FDSAlertWebhookURL: "https://hooks.slack.com/test",
					},
					Environment: constant.EnvironmentStaging,
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
				WithCustomerRepository(customerRepo),
				WithRabbitMqExt(rabbitMq),
			)

			// Create context with trace ID for testing
			ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, "test-trace-id")

			// Call the function
			resp, err := service.CheckTransaction(ctx, transactionID, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp, "Response should not be nil when no error is expected")
				if tc.expectedResp != nil && resp != nil {
					assert.Equal(t, tc.expectedResp.Status, resp.Status)
					assert.True(t, tc.expectedResp.Score.Equal(resp.Score), "Expected score %s, got %s", tc.expectedResp.Score.String(), resp.Score.String())
				}
			}

			// Verify all expectations were met
			accountTrxRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			fraudRulesRepo.AssertExpectations(t)
			ruleEvalRepo.AssertExpectations(t)
			fraudNetMock.AssertExpectations(t)
		})
	}
}
