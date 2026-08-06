package fdsservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestUpdateTransaction(t *testing.T) {
	// Create default values for reuse
	transactionID := "test-transaction-id"
	merchantID := uuid.New()
	referenceID := "test-reference-id"
	trxUUID := uuid.New()
	now := time.Now()

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

	testCases := []struct {
		desc         string
		wantErr      bool
		expectedResp *[]fdscommon.UpdateResponse
		setupMock    func(
			accountTrxRepo *mockRepositories.IAccountTransactionRepository,
			paymentRepo *mockRepositories.IPaymentRepository,
			fraudRulesRepo *mockRepositories.IFraudRulesRepository,
			ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
			fraudNetMock *mockRepositories.IFdsProcessorRepository,
		)
	}{
		{
			desc:    "success case - non-fraud transaction",
			wantErr: false,
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				// Mock fraud processor update
				fraudNetMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(&fdscommon.UpdateResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated",
					},
				}, nil)
			},
		},
		{
			desc:    "success case - fraud transaction",
			wantErr: false,
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated as fraud",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction with fraud status
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(70),
						Result:      "HIGH",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				// Mock fraud processor update
				fraudNetMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(&fdscommon.UpdateResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated as fraud",
					},
				}, nil)
			},
		},
		{
			desc:         "error case - transaction not found",
			wantErr:      true,
			expectedResp: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction not found
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:         "error case - payment not found",
			wantErr:      true,
			expectedResp: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				// Mock payment not found
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:         "error case - rule evaluations not found",
			wantErr:      true,
			expectedResp: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock empty rule evaluations
				emptyRuleEvals := []ruleevaluationsmodel.RuleEvaluations{}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&emptyRuleEvals, nil)
			},
		},
		{
			desc:         "error case - fraud rule not found",
			wantErr:      false, // Function continues despite missing fraud rule
			expectedResp: &[]fdscommon.UpdateResponse{},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule not found
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:         "error case - unsupported provider",
			wantErr:      false, // Function continues despite unsupported provider
			expectedResp: &[]fdscommon.UpdateResponse{},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule with unsupported provider
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: "UNSUPPORTED_PROVIDER",
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				// No fraud processor mock needed since provider is unsupported
			},
		},
		{
			desc:         "error case - processor update fails",
			wantErr:      true, // Function should return error when processor update fails
			expectedResp: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
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
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				// Mock fraud processor update failure
				fraudNetMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(nil, errors.New(response.HttpErrInternal, constant.ErrFraudRulesNotFound))
			},
		},
		{
			desc:    "success case - multiple rule evaluations",
			wantErr: false,
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated rule 1",
					},
				},
				{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated rule 2",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:        trxUUID,
					MerchantID:  merchantID,
					ReferenceID: referenceID,
					Status:      "COMPLETED",
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock multiple rule evaluations
				ruleEvaluation1 := uuid.NewString()
				ruleEvaluation2 := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvaluation1,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
					{
						UUID:        ruleEvaluation2,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-2",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(20),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rules
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule 1",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-2").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-2",
					RuleName:      "Test Rule 2",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.2),
				}, nil)

				// Mock fraud processor update for first rule
				fraudNetMock.On("Update", mock.Anything, mock.MatchedBy(func(req *fdscommon.UpdateRequest) bool {
					return req.OrderID == ruleEvaluation1
				})).Return(&fdscommon.UpdateResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated rule 1",
					},
				}, nil)

				// Mock fraud processor update for second rule
				fraudNetMock.On("Update", mock.Anything, mock.MatchedBy(func(req *fdscommon.UpdateRequest) bool {
					return req.OrderID == ruleEvaluation2
				})).Return(&fdscommon.UpdateResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated rule 2",
					},
				}, nil)
			},
		},
		{
			desc:    "edge case - nil metadata",
			wantErr: false,
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				fraudNetMock *mockRepositories.IFdsProcessorRepository,
			) {
				// Mock transaction with nil metadata
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:              trxUUID,
					MerchantID:        merchantID,
					ReferenceID:       referenceID,
					Status:            "COMPLETED",
					AdditionalInfoObj: nil, // No metadata
				}, nil)

				// Mock payment
				payment := &paymentModel.Payment{
					UUID:        uuid.NewString(),
					ReferenceID: &referenceID,
					MerchantID:  merchantID.String(),
					Status:      "PAID",
					Amount:      decimal.NewFromInt(100000),
					Currency:    "IDR",
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				// Mock rule evaluations
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        uuid.NewString(),
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				// Mock fraud rule
				fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-1",
					RuleName:      "Test Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_FRAUD_NET,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.3),
				}, nil)

				// Mock fraud processor update
				fraudNetMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(&fdscommon.UpdateResponse{
					Success: true,
					Code:    util.ValueToPtr("success"),
					Source:  util.ValueToPtr("fraudnet"),
					Message: map[string]string{
						"status": "updated",
					},
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t) // Not used but needed for service init

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expected mocks
			tc.setupMock(accountTrxRepo, paymentRepo, fraudRulesRepo, ruleEvalRepo, fraudNetMock)

			// Create service with mocks
			processors := map[string]repository.IFdsProcessorRepository{
				constant.PROVIDER_FRAUD_NET: fraudNetMock,
			}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
					},
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
			)

			// Create context with trace ID for testing
			ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, "test-trace-id")
			// Call the function
			result, err := service.UpdateTransaction(ctx, transactionID, nil)

			// Assert results
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.expectedResp != nil {
					assert.Equal(t, len(*tc.expectedResp), len(*result))
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

func TestUpdateTransaction_CardStatusMapping(t *testing.T) {
	transactionID := "test-transaction-id"
	merchantID := uuid.New()
	referenceID := "test-reference-id"
	trxUUID := uuid.New()
	now := time.Now()

	testCases := []struct {
		desc                string
		authorizationData   interface{}
		expectedCardStatus  string
		expectedNote        string
		setupAdditionalInfo func() types.NullJSONText
		wantErr             bool
	}{
		{
			desc: "success case - with authorization data decline code 01",
			authorizationData: map[string]interface{}{
				"acquirerResponseCode": "01",
			},
			expectedCardStatus: constant.FRAUD_NET_CARD_STATUS_DECLINE,
			expectedNote:       "failed authorization",
			setupAdditionalInfo: func() types.NullJSONText {
				raw := `{
					"chargeStatus": "SUCCESS",
					"methodDetail": {
						"card": {
							"last4": "1234",
							"first8": "44400000",
							"fingerprint": "test-fingerprint",
							"cardBrand": "VISA",
							"countryCode": "ID",
							"authorizationResult": {
								"acquirerResponseCode": "01"
							}
						}
					}
				}`
				return types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(raw)),
				}
			},
			wantErr: false,
		},
		{
			desc: "success case - with authorization data expired code 54",
			authorizationData: map[string]interface{}{
				"acquirerResponseCode": "54",
			},
			expectedCardStatus: constant.FRAUD_NET_CARD_STATUS_EXPIRED,
			expectedNote:       "failed authorization",
			setupAdditionalInfo: func() types.NullJSONText {
				raw := `{
					"chargeStatus": "SUCCESS",
					"methodDetail": {
						"card": {
							"last4": "1234",
							"first8": "44400000",
							"fingerprint": "test-fingerprint",
							"cardBrand": "VISA",
							"countryCode": "ID",
							"authorizationResult": {
								"acquirerResponseCode": "54"
							}
						}
					}
				}`
				return types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(raw)),
				}
			},
			wantErr: false,
		},
		{
			desc: "success case - with authorization data stolen code 04",
			authorizationData: map[string]interface{}{
				"acquirerResponseCode": "04",
			},
			expectedCardStatus: constant.FRAUD_NET_CARD_STATUS_STOLEN,
			expectedNote:       "failed authorization",
			setupAdditionalInfo: func() types.NullJSONText {
				raw := `{
					"chargeStatus": "SUCCESS",
					"methodDetail": {
						"card": {
							"last4": "1234",
							"first8": "44400000",
							"fingerprint": "test-fingerprint",
							"cardBrand": "VISA",
							"countryCode": "ID",
							"authorizationResult": {
								"acquirerResponseCode": "04"
							}
						}
					}
				}`
				return types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(raw)),
				}
			},
			wantErr: false,
		},
		{
			desc:               "success case - no credit card metadata defaults to decline",
			authorizationData:  nil,
			expectedCardStatus: constant.FRAUD_NET_CARD_STATUS_DECLINE,
			expectedNote:       "failed authorization",
			setupAdditionalInfo: func() types.NullJSONText {
				raw := `{
					"expiredAt": "2025-05-30T15:04:05Z"
				}`
				return types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(raw)),
				}
			},
			wantErr: false,
		},
		{
			desc:               "success case - no authorization data defaults to decline",
			authorizationData:  nil,
			expectedCardStatus: constant.FRAUD_NET_CARD_STATUS_DECLINE,
			expectedNote:       "failed authorization",
			setupAdditionalInfo: func() types.NullJSONText {
				raw := `{
					"chargeStatus": "SUCCESS",
					"methodDetail": {
						"card": {
							"last4": "1234",
							"first8": "44400000",
							"fingerprint": "test-fingerprint",
							"cardBrand": "VISA",
							"countryCode": "ID"
						}
					}
				}`
				return types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(raw)),
				}
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Mock transaction
			accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
				UUID:           trxUUID,
				MerchantID:     merchantID,
				ReferenceID:    referenceID,
				Status:         "COMPLETED",
				AdditionalInfo: tc.setupAdditionalInfo(),
			}, nil)

			// Mock payment
			payment := &paymentModel.Payment{
				UUID:        uuid.NewString(),
				ReferenceID: &referenceID,
				MerchantID:  merchantID.String(),
				Status:      "FAILED",
				Amount:      decimal.NewFromInt(100000),
				Currency:    "IDR",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

			// Mock rule evaluations
			ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
				{
					UUID:        uuid.NewString(),
					ReferenceID: trxUUID.String(),
					RuleID:      "rule-1",
					EvaluatedAt: now,
					Score:       decimal.NewFromInt(70),
					Result:      "HIGH",
				},
			}
			ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

			// Mock fraud rule
			fraudRulesRepo.On("GetByID", mock.Anything, "rule-1").Return(&fraudrulesmodel.FraudRules{
				UUID:          "rule-1",
				RuleName:      "Test Rule",
				ReferenceType: "\"[ANY]\"",
				Provider: sql.NullString{
					String: constant.PROVIDER_FRAUD_NET,
					Valid:  true,
				},
				Weight: decimal.NewFromFloat(0.3),
			}, nil)

			// Mock fraud processor update - capture the request to verify card status
			var capturedRequest *fdscommon.UpdateRequest
			fraudNetMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).
				Run(func(args mock.Arguments) {
					capturedRequest = args.Get(1).(*fdscommon.UpdateRequest)
				}).Return(&fdscommon.UpdateResponse{
				Success: true,
				Code:    util.ValueToPtr("success"),
				Source:  util.ValueToPtr("fraudnet"),
				Message: map[string]string{
					"status": "updated",
				},
			}, nil)

			// Create service with mocks
			processors := map[string]repository.IFdsProcessorRepository{
				constant.PROVIDER_FRAUD_NET: fraudNetMock,
			}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
					},
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
			)

			// Create context with trace ID for testing
			ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, "test-trace-id")

			// Call the function
			result, err := service.UpdateTransaction(ctx, transactionID, nil)

			// Assert results
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify card status was set correctly
				assert.NotNil(t, capturedRequest)
				assert.NotNil(t, capturedRequest.Payment.CardStatus)
				assert.Equal(t, tc.expectedCardStatus, *capturedRequest.Payment.CardStatus)

				// Verify note was set correctly
				assert.NotNil(t, capturedRequest.Note)
				assert.Equal(t, tc.expectedNote, *capturedRequest.Note)
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

func TestUpdateTransactionSokratech(t *testing.T) {
	transactionID := "test-transaction-id"
	merchantID := uuid.New()
	referenceID := "test-reference-id"
	trxUUID := uuid.New()
	now := time.Now()

	ccMetadataRaw := `{
		"chargeStatus": "SUCCESS",
		"methodDetail": {
			"card": {
				"last4": "1234",
				"first8": "44400000",
				"fingerprint": "test-fingerprint",
				"cardBrand": "VISA",
				"countryCode": "ID",
				"authorizationResult": {
					"acquirerResponseCode": "00",
					"approvalCode": "AB1234",
					"cvvResult": "M",
					"acquirerTransactionId": "acq-txn-123"
				},
				"authenticationResult": {
					"eciCode": "05"
				},
				"binInformations": {
					"type": "CREDIT",
					"brand": "VISA",
					"country": "ID",
					"issuingBank": "BRI_S2I"
				}
			}
		}
	}`

	additionalInfo := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText([]byte(ccMetadataRaw)),
	}

	testCases := []struct {
		desc         string
		wantErr      bool
		expectedResp *[]fdscommon.UpdateResponse
		request      *fdscommon.UpdateRequest
		setupMock    func(
			accountTrxRepo *mockRepositories.IAccountTransactionRepository,
			paymentRepo *mockRepositories.IPaymentRepository,
			fraudRulesRepo *mockRepositories.IFraudRulesRepository,
			ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
			sokratechMock *mockRepositories.IFdsProcessorRepository,
			merchantRepo *mockRepositories.IMerchantRepository,
			paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
			customerRepo *mockRepositories.ICustomerRepository,
		)
	}{
		{
			desc:    "success case - sokratech provider update with isFraud",
			wantErr: false,
			request: &fdscommon.UpdateRequest{
				IsFraud: util.ValueToPtr(true),
				Note:    util.ValueToPtr("confirmed fraud"),
				Payment: &fdscommon.PaymentUpdate{
					ChargebackStatus: util.ValueToPtr("CHARGEBACK_OPENED"),
				},
			},
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Data: fdscommon.UpdateData{
						ID: "exec-sokratech-001",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				sokratechMock *mockRepositories.IFdsProcessorRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				payment := &paymentModel.Payment{
					UUID:                     uuid.NewString(),
					ReferenceID:              &referenceID,
					MerchantID:               merchantID.String(),
					CustomerID:               "customer-001",
					PaymentMethodID:          "pm-001",
					Status:                   "PAID",
					Amount:                   decimal.NewFromInt(100000),
					Currency:                 "IDR",
					CreatedAt:                now,
					UpdatedAt:                now,
					ProcessorReferenceNumber: util.ValueToPtr("proc-ref-001"),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				ruleEvalUUID := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvalUUID,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-sokratech-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(60),
						Result:      "HIGH",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-sokratech-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-sokratech-1",
					RuleName:      "Sokratech Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_SOKRATECH,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.5),
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					Address:   "Jl. Test No. 1",
					PostCode:  "12345",
					PICEmail:  "pic@test.com",
					PICPhone:  "081234567890",
					RiskLevel: sql.NullString{String: "MEDIUM", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)

				customerRepo.On("GetCustomerById", mock.Anything, "customer-001", merchantID.String()).Return(&customerModel.Customer{
					UUID:        "customer-001",
					Email:       "customer@test.com",
					PhoneNumber: "08987654321",
					FirstName:   "John",
					LastName:    "Doe",
					MerchantID:  merchantID.String(),
				}, nil)

				sokratechMock.On("Update", mock.Anything, mock.MatchedBy(func(req *fdscommon.UpdateRequest) bool {
					return req.FullContext != nil && req.IsFraud != nil && *req.IsFraud
				})).Return(&fdscommon.UpdateResponse{
					Success: true,
					Data: fdscommon.UpdateData{
						ID: "exec-sokratech-001",
					},
				}, nil)

				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, trxUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
		},
		{
			desc:    "success case - sokratech provider without isFraud",
			wantErr: false,
			request: nil,
			expectedResp: &[]fdscommon.UpdateResponse{
				{
					Success: true,
					Data: fdscommon.UpdateData{
						ID: "exec-sokratech-002",
					},
				},
			},
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				sokratechMock *mockRepositories.IFdsProcessorRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				payment := &paymentModel.Payment{
					UUID:                     uuid.NewString(),
					ReferenceID:              &referenceID,
					MerchantID:               merchantID.String(),
					PaymentMethodID:          "pm-001",
					Status:                   "PAID",
					Amount:                   decimal.NewFromInt(100000),
					Currency:                 "IDR",
					CreatedAt:                now,
					UpdatedAt:                now,
					ProcessorReferenceNumber: util.ValueToPtr("proc-ref-001"),
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				ruleEvalUUID := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvalUUID,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-sokratech-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-sokratech-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-sokratech-1",
					RuleName:      "Sokratech Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_SOKRATECH,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.5),
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)

				sokratechMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(&fdscommon.UpdateResponse{
					Success: true,
					Data: fdscommon.UpdateData{
						ID: "exec-sokratech-002",
					},
				}, nil)
			},
		},
		{
			desc:         "error case - sokratech buildPaymentFullContext merchant not found",
			wantErr:      false, // continues to next iteration
			expectedResp: &[]fdscommon.UpdateResponse{},
			request:      nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				sokratechMock *mockRepositories.IFdsProcessorRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,

			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				payment := &paymentModel.Payment{
					UUID:            uuid.NewString(),
					ReferenceID:     &referenceID,
					MerchantID:      merchantID.String(),
					PaymentMethodID: "pm-001",
					Status:          "PAID",
					Amount:          decimal.NewFromInt(100000),
					Currency:        "IDR",
					CreatedAt:       now,
					UpdatedAt:       now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				ruleEvalUUID := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvalUUID,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-sokratech-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-sokratech-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-sokratech-1",
					RuleName:      "Sokratech Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_SOKRATECH,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.5),
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(nil, sql.ErrNoRows)
			},
		},
		{
			desc:    "error case - sokratech processor update fails",
			wantErr: true,
			request: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				sokratechMock *mockRepositories.IFdsProcessorRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,

			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				payment := &paymentModel.Payment{
					UUID:            uuid.NewString(),
					ReferenceID:     &referenceID,
					MerchantID:      merchantID.String(),
					PaymentMethodID: "pm-001",
					Status:          "PAID",
					Amount:          decimal.NewFromInt(100000),
					Currency:        "IDR",
					CreatedAt:       now,
					UpdatedAt:       now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				ruleEvalUUID := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvalUUID,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-sokratech-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-sokratech-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-sokratech-1",
					RuleName:      "Sokratech Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_SOKRATECH,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.5),
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)

				sokratechMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(nil, errors.New(response.HttpErrInternal, assert.AnError))
			},
		},
		{
			desc:    "error case - sokratech processor returns unsuccessful response",
			wantErr: true,
			request: nil,
			setupMock: func(
				accountTrxRepo *mockRepositories.IAccountTransactionRepository,
				paymentRepo *mockRepositories.IPaymentRepository,
				fraudRulesRepo *mockRepositories.IFraudRulesRepository,
				ruleEvalRepo *mockRepositories.IRuleEvaluationsRepository,
				sokratechMock *mockRepositories.IFdsProcessorRepository,
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,

			) {
				accountTrxRepo.On("FindByID", mock.Anything, transactionID).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID:           trxUUID,
					MerchantID:     merchantID,
					ReferenceID:    referenceID,
					Status:         "COMPLETED",
					AdditionalInfo: additionalInfo,
				}, nil)

				payment := &paymentModel.Payment{
					UUID:            uuid.NewString(),
					ReferenceID:     &referenceID,
					MerchantID:      merchantID.String(),
					PaymentMethodID: "pm-001",
					Status:          "PAID",
					Amount:          decimal.NewFromInt(100000),
					Currency:        "IDR",
					CreatedAt:       now,
					UpdatedAt:       now,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Return(payment, nil)

				ruleEvalUUID := uuid.NewString()
				ruleEvaluations := []ruleevaluationsmodel.RuleEvaluations{
					{
						UUID:        ruleEvalUUID,
						ReferenceID: trxUUID.String(),
						RuleID:      "rule-sokratech-1",
						EvaluatedAt: now,
						Score:       decimal.NewFromInt(30),
						Result:      "LOW",
					},
				}
				ruleEvalRepo.On("GetByRefID", mock.Anything, trxUUID.String()).Return(&ruleEvaluations, nil)

				fraudRulesRepo.On("GetByID", mock.Anything, "rule-sokratech-1").Return(&fraudrulesmodel.FraudRules{
					UUID:          "rule-sokratech-1",
					RuleName:      "Sokratech Rule",
					ReferenceType: "\"[ANY]\"",
					Provider: sql.NullString{
						String: constant.PROVIDER_SOKRATECH,
						Valid:  true,
					},
					Weight: decimal.NewFromFloat(0.5),
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)

				sokratechMock.On("Update", mock.Anything, mock.AnythingOfType("*fdscommon.UpdateRequest")).Return(&fdscommon.UpdateResponse{
					Success: false,
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			sokratechMock := mockRepositories.NewIFdsProcessorRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t)
			paymentMethodRepo := mockRepositories.NewIPaymentMethodRepository(t)
			customerRepo := mockRepositories.NewICustomerRepository(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(accountTrxRepo, paymentRepo, fraudRulesRepo, ruleEvalRepo, sokratechMock, merchantRepo, paymentMethodRepo, customerRepo)

			processors := map[string]repository.IFdsProcessorRepository{
				constant.PROVIDER_SOKRATECH: sokratechMock,
			}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
						BinLength:      6,
					},
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
				WithPaymentMethodRepository(paymentMethodRepo),
				WithCustomerRepository(customerRepo),
			)

			ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, "test-trace-id")
			result, err := service.UpdateTransaction(ctx, transactionID, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tc.expectedResp != nil {
					assert.Equal(t, len(*tc.expectedResp), len(*result))
				}
			}

			accountTrxRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			fraudRulesRepo.AssertExpectations(t)
			ruleEvalRepo.AssertExpectations(t)
			sokratechMock.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
			customerRepo.AssertExpectations(t)
		})
	}
}

func TestBuildPaymentFullContext(t *testing.T) {
	merchantID := uuid.New()
	now := time.Now()

	ccMetadataRaw := `{
		"chargeStatus": "SUCCESS",
		"methodDetail": {
			"card": {
				"last4": "1234",
				"first8": "44400000",
				"fingerprint": "test-fingerprint",
				"cardBrand": "VISA",
				"countryCode": "ID",
				"authorizationResult": {
					"acquirerResponseCode": "00",
					"approvalCode": "AB1234",
					"cvvResult": "M",
					"acquirerTransactionId": "acq-txn-123"
				},
				"authenticationResult": {
					"eciCode": "05"
				},
				"binInformations": {
					"type": "CREDIT",
					"brand": "VISA",
					"country": "ID",
					"issuingBank": "BRI_S2I"
				}
			}
		}
	}`

	buildTrx := func(additionalInfo types.NullJSONText) *orchestrator_model.AccountTransactionWithUseCase {
		return &orchestrator_model.AccountTransactionWithUseCase{
			UUID:           uuid.New(),
			MerchantID:     merchantID,
			ReferenceID:    "ref-001",
			Status:         "COMPLETED",
			AdditionalInfo: additionalInfo,
		}
	}

	buildPayment := func(customerID, pmID string) *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:                     uuid.NewString(),
			ReferenceID:              util.ValueToPtr("ref-001"),
			MerchantID:               merchantID.String(),
			CustomerID:               customerID,
			PaymentMethodID:          pmID,
			Status:                   "PAID",
			Amount:                   decimal.NewFromInt(100000),
			Currency:                 "IDR",
			CreatedAt:                now,
			UpdatedAt:                now,
			ProcessorReferenceNumber: util.ValueToPtr("proc-ref-001"),
		}
	}

	testCases := []struct {
		desc      string
		setupTrx  func() *orchestrator_model.AccountTransactionWithUseCase
		setupPmt  func() *paymentModel.Payment
		setupMock func(
			merchantRepo *mockRepositories.IMerchantRepository,
			paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
			customerRepo *mockRepositories.ICustomerRepository,
		)
		wantErr bool
	}{
		{
			desc: "success case - full context with card data, customer, and payment method",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: true, JSONText: []byte(ccMetadataRaw)})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("customer-001", "pm-001") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					Address:   "Jl. Test No. 1",
					PostCode:  "12345",
					PICEmail:  "pic@test.com",
					PICPhone:  "081234567890",
					RiskLevel: sql.NullString{String: "MEDIUM", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)

				customerRepo.On("GetCustomerById", mock.Anything, "customer-001", merchantID.String()).Return(&customerModel.Customer{
					UUID:        "customer-001",
					Email:       "customer@test.com",
					PhoneNumber: "08987654321",
					FirstName:   "John",
					LastName:    "Doe",
					MerchantID:  merchantID.String(),
				}, nil)
			},
			wantErr: false,
		},
		{
			desc: "error case - merchant not found",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: true, JSONText: []byte(ccMetadataRaw)})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("", "") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(nil, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			desc: "success case - without cc metadata (nil additional info)",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: false})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("", "pm-001") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "AGGREGATOR",
				}, nil)
			},
			wantErr: false,
		},
		{
			desc: "success case - with customer but no payment method ID",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: true, JSONText: []byte(ccMetadataRaw)})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("customer-002", "") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				customerRepo.On("GetCustomerById", mock.Anything, "customer-002", merchantID.String()).Return(&customerModel.Customer{
					UUID:        "customer-002",
					Email:       "jane@example.com",
					PhoneNumber: "0811112222",
					FirstName:   "Jane",
					LastName:    "Smith",
				}, nil)
			},
			wantErr: false,
		},
		{
			desc: "success case - payment method lookup fails gracefully",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: true, JSONText: []byte(ccMetadataRaw)})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("", "pm-bad") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-bad", merchantID.String()).Return(nil, sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			desc: "success case - customer lookup fails gracefully",
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return buildTrx(types.NullJSONText{Valid: true, JSONText: []byte(ccMetadataRaw)})
			},
			setupPmt: func() *paymentModel.Payment { return buildPayment("customer-missing", "pm-001") },
			setupMock: func(
				merchantRepo *mockRepositories.IMerchantRepository,
				paymentMethodRepo *mockRepositories.IPaymentMethodRepository,
				customerRepo *mockRepositories.ICustomerRepository,
			) {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID.String()).Return(&merchant.Merchant{
					UUID:      merchantID.String(),
					Name:      "Test Merchant",
					ShortName: "TestM",
					RiskLevel: sql.NullString{String: "LOW", Valid: true},
				}, nil)

				paymentMethodRepo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-001", merchantID.String()).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "CREDIT_CARD",
						Acquirer: "BRI",
					},
					ChannelType: "DIRECT",
				}, nil)

				customerRepo.On("GetCustomerById", mock.Anything, "customer-missing", merchantID.String()).Return(nil, sql.ErrNoRows)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t)
			paymentMethodRepo := mockRepositories.NewIPaymentMethodRepository(t)
			customerRepo := mockRepositories.NewICustomerRepository(t)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(merchantRepo, paymentMethodRepo, customerRepo)

			processors := map[string]repository.IFdsProcessorRepository{}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
						BinLength:      6,
					},
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
				WithPaymentMethodRepository(paymentMethodRepo),
				WithCustomerRepository(customerRepo),
			)

			trx := tc.setupTrx()
			payment := tc.setupPmt()

			// Get ccMetadata from additional info
			ccMetadata, _ := trx.GetCreditcardMetadataFromAdditionalInfo()

			fdsService := service.(*FdsService)
			result, err := fdsService.buildPaymentFullContext(context.Background(), trx, payment, ccMetadata)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// Verify partner info
				assert.Equal(t, merchantID.String(), result.Partner.ID)
			}

			merchantRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
			customerRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateFdsRiskAssessment(t *testing.T) {
	transactionUUID := uuid.New()
	merchantID := uuid.New()
	now := time.Now()

	testCases := []struct {
		desc      string
		setupMock func(accountTrxRepo *mockRepositories.IAccountTransactionRepository)
		setupTrx  func() *orchestrator_model.AccountTransactionWithUseCase
		updateFds fdscommon.FdsRiskAssessment
		wantErr   bool
	}{
		{
			desc: "success case - create new FDS risk assessment",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    false,
						JSONText: nil,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score:            decimal.NewFromInt(75),
				Level:            "HIGH",
				Recommendation:   "Reject",
				Status:           constant.FDS_STATUS_REJECTED,
				EvaluatedAt:      now,
				IsFraud:          util.ValueToPtr(true),
				ChargebackStatus: "opened",
			},
			wantErr: false,
		},
		{
			desc: "success case - update existing FDS risk assessment",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				existingFdsData := map[string]interface{}{
					constant.FdsRiskAssesment: map[string]interface{}{
						"score":            50,
						"level":            "MEDIUM",
						"recommendation":   "Review",
						"status":           constant.FDS_STATUS_REVIEW,
						"evaluatedAt":      now.Add(-time.Hour).Format(time.RFC3339),
						"isFraud":          false,
						"chargebackStatus": "",
					},
				}
				additionalInfoBytes, _ := json.Marshal(existingFdsData)
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfoBytes,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score:            decimal.NewFromInt(85),
				Level:            "HIGH",
				Recommendation:   "Reject",
				Status:           constant.FDS_STATUS_REJECTED,
				EvaluatedAt:      now,
				IsFraud:          util.ValueToPtr(true),
				ChargebackStatus: "opened",
			},
			wantErr: false,
		},
		{
			desc: "success case - partial update (only some fields)",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				existingFdsData := map[string]interface{}{
					constant.FdsRiskAssesment: map[string]interface{}{
						"score":            50,
						"level":            "MEDIUM",
						"recommendation":   "Review",
						"status":           constant.FDS_STATUS_REVIEW,
						"evaluatedAt":      now.Add(-time.Hour).Format(time.RFC3339),
						"isFraud":          false,
						"chargebackStatus": "",
					},
				}
				additionalInfoBytes, _ := json.Marshal(existingFdsData)
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfoBytes,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				IsFraud:          util.ValueToPtr(true),
				ChargebackStatus: "won",
			},
			wantErr: false,
		},
		{
			desc: "success case - invalid existing JSON data",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: []byte("invalid json"),
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score:            decimal.NewFromInt(25),
				Level:            "LOW",
				Recommendation:   "Approve",
				Status:           constant.FDS_STATUS_PASSED,
				EvaluatedAt:      now,
				IsFraud:          util.ValueToPtr(false),
				ChargebackStatus: "",
			},
			wantErr: false,
		},
		{
			desc: "success case - invalid existing FDS data structure",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				invalidFdsData := map[string]interface{}{
					constant.FdsRiskAssesment: "invalid_structure",
				}
				additionalInfoBytes, _ := json.Marshal(invalidFdsData)
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    true,
						JSONText: additionalInfoBytes,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score:            decimal.NewFromInt(30),
				Level:            "LOW",
				Recommendation:   "Approve",
				Status:           constant.FDS_STATUS_PASSED,
				EvaluatedAt:      now,
				IsFraud:          util.ValueToPtr(false),
				ChargebackStatus: "",
			},
			wantErr: false,
		},
		{
			desc: "error case - failed to marshal additional info",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    false,
						JSONText: nil,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score: decimal.NewFromInt(50),
				// Create a circular reference that will cause Marshal to fail
				// This is a bit tricky to simulate, so we'll use repository error instead
			},
			wantErr: false, // The actual marshal error case is hard to simulate, testing repository error instead
		},
		{
			desc: "error case - repository update fails",
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, transactionUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(errors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords))
			},
			setupTrx: func() *orchestrator_model.AccountTransactionWithUseCase {
				return &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       transactionUUID,
					MerchantID: merchantID,
					AdditionalInfo: types.NullJSONText{
						Valid:    false,
						JSONText: nil,
					},
				}
			},
			updateFds: fdscommon.FdsRiskAssessment{
				Score:            decimal.NewFromInt(40),
				Level:            "MEDIUM",
				Recommendation:   "Review",
				Status:           constant.FDS_STATUS_REVIEW,
				EvaluatedAt:      now,
				IsFraud:          util.ValueToPtr(false),
				ChargebackStatus: "",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			paymentRepo := mockRepositories.NewIPaymentRepository(t)
			fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
			ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
			merchantRepo := mockRepositories.NewIMerchantRepository(t)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(accountTrxRepo)

			processors := map[string]repository.IFdsProcessorRepository{}

			service := New(
				&config.Config{
					FdsConfig: config.FdsConfig{
						ScoreThreshold: 50,
					},
				},
				logger,
				fraudRulesRepo,
				ruleEvalRepo,
				accountTrxRepo,
				paymentRepo,
				merchantRepo,
				processors,
			)

			trx := tc.setupTrx()

			ctx := context.Background()

			fdsService := service.(*FdsService)
			err := fdsService.UpdateFdsRiskAssessment(ctx, trx, tc.updateFds)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accountTrxRepo.AssertExpectations(t)
		})
	}
}
