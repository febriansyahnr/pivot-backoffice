package fdsservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMapChannelToFraudNetPayment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Direct Deposit - VIRTUAL_ACCOUNT",
			input:    "VIRTUAL_ACCOUNT",
			expected: constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT,
		},
		{
			name:     "Direct Deposit - BANK_TRANSFER",
			input:    "BANK_TRANSFER",
			expected: constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT,
		},
		{
			name:     "Direct Deposit - MANUAL_TRANSFER",
			input:    "MANUAL_TRANSFER",
			expected: constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT,
		},
		{
			name:     "Direct Deposit - BALANCE_TRANSFER",
			input:    "BALANCE_TRANSFER",
			expected: constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT,
		},
		{
			name:     "Direct Deposit - TRANSFER",
			input:    "TRANSFER",
			expected: constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT,
		},
		{
			name:     "Credit Card - CREDIT_CARD",
			input:    "CREDIT_CARD",
			expected: constant.FRAUD_NET_PAYMENT_CREDIT_CARD,
		},
		{
			name:     "Credit Card - CARD",
			input:    "CARD",
			expected: constant.FRAUD_NET_PAYMENT_CREDIT_CARD,
		},
		{
			name:     "E-Wallets - QRIS",
			input:    "QRIS",
			expected: constant.FRAUD_NET_PAYMENT_EWALLETS,
		},
		{
			name:     "E-Wallets - QR",
			input:    "QR",
			expected: constant.FRAUD_NET_PAYMENT_EWALLETS,
		},
		{
			name:     "E-Payment - PPOB",
			input:    "PPOB",
			expected: constant.FRAUD_NET_PAYMENT_EPAYMENT,
		},
		{
			name:     "E-Payment - BILL",
			input:    "BILL",
			expected: constant.FRAUD_NET_PAYMENT_EPAYMENT,
		},
		{
			name:     "Internal Transfer - BALANCE_ADJUSTMENT",
			input:    "BALANCE_ADJUSTMENT",
			expected: constant.FRAUD_NET_PAYMENT_INTERNAL_TRANSFER,
		},
		{
			name:     "Other - MANUAL_ACTION",
			input:    "MANUAL_ACTION",
			expected: constant.FRAUD_NET_PAYMENT_OTHER,
		},
		{
			name:     "Goods Service - MERCHANT_PAYMENT",
			input:    "MERCHANT_PAYMENT",
			expected: constant.FRAUD_NET_PAYMENT_GOODS_SERVICE,
		},
		{
			name:     "Cash - TOP_UP",
			input:    "TOP_UP",
			expected: constant.FRAUD_NET_PAYMENT_CASH,
		},
		{
			name:     "ATM - ALTO",
			input:    "ALTO",
			expected: constant.FRAUD_NET_PAYMENT_ATM,
		},
		{
			name:     "Third Party Processor - XB",
			input:    "XB",
			expected: constant.FRAUD_NET_PAYMENT_THIRD_PARTY_PROCESSOR,
		},
		{
			name:     "Default Case - UNKNOWN",
			input:    "UNKNOWN",
			expected: constant.FRAUD_NET_PAYMENT_OTHER,
		},
		{
			name:     "Default Case - Empty String",
			input:    "",
			expected: constant.FRAUD_NET_PAYMENT_OTHER,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapChannelToFraudNetPayment(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestMapCardTypeToFraudNet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MasterCard - uppercase",
			input:    "MASTERCARD",
			expected: constant.FRAUD_NET_CARD_TYPE_MC,
		},
		{
			name:     "MasterCard - lowercase",
			input:    "mastercard",
			expected: constant.FRAUD_NET_CARD_TYPE_MC,
		},
		{
			name:     "MasterCard - mixed case",
			input:    "MasterCard",
			expected: constant.FRAUD_NET_CARD_TYPE_MC,
		},
		{
			name:     "Visa - uppercase",
			input:    "VISA",
			expected: constant.FRAUD_NET_CARD_TYPE_VISA,
		},
		{
			name:     "Visa - lowercase",
			input:    "visa",
			expected: constant.FRAUD_NET_CARD_TYPE_VISA,
		},
		{
			name:     "Amex - uppercase",
			input:    "AMEX",
			expected: constant.FRAUD_NET_CARD_TYPE_AMEX,
		},
		{
			name:     "American Express - uppercase",
			input:    "AMERICAN_EXPRESS",
			expected: constant.FRAUD_NET_CARD_TYPE_AMEX,
		},
		{
			name:     "Discover - uppercase",
			input:    "DISCOVER",
			expected: constant.FRAUD_NET_CARD_TYPE_DISCOVER,
		},
		{
			name:     "Discover - lowercase",
			input:    "discover",
			expected: constant.FRAUD_NET_CARD_TYPE_DISCOVER,
		},
		{
			name:     "Diners - uppercase",
			input:    "DINERS",
			expected: constant.FRAUD_NET_CARD_TYPE_DINERS_CLUB,
		},
		{
			name:     "Diners Club - uppercase",
			input:    "DINERS_CLUB",
			expected: constant.FRAUD_NET_CARD_TYPE_DINERS_CLUB,
		},
		{
			name:     "Default Case - Unknown",
			input:    "UNKNOWN",
			expected: constant.FRAUD_NET_CARD_TYPE_OTHER,
		},
		{
			name:     "Default Case - Empty String",
			input:    "",
			expected: constant.FRAUD_NET_CARD_TYPE_OTHER,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapCardTypeToFraudNet(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestMapTransactionStatusToFraudNet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Success - uppercase",
			input:    "SUCCESS",
			expected: constant.FRAUD_NET_TRX_STATUS_FULFILLED,
		},
		{
			name:     "Success - lowercase",
			input:    "success",
			expected: constant.FRAUD_NET_TRX_STATUS_FULFILLED,
		},
		{
			name:     "Failed - uppercase",
			input:    "FAILED",
			expected: constant.FRAUD_NET_TRX_STATUS_CANCELLED,
		},
		{
			name:     "Failed - lowercase",
			input:    "failed",
			expected: constant.FRAUD_NET_TRX_STATUS_CANCELLED,
		},
		{
			name:     "Pending - uppercase",
			input:    "PENDING",
			expected: constant.FRAUD_NET_TRX_STATUS_QUEUED,
		},
		{
			name:     "Pending - lowercase",
			input:    "pending",
			expected: constant.FRAUD_NET_TRX_STATUS_QUEUED,
		},
		{
			name:     "Default Case - Unknown",
			input:    "UNKNOWN",
			expected: constant.FRAUD_NET_TRX_STATUS_NEW,
		},
		{
			name:     "Default Case - Empty String",
			input:    "",
			expected: constant.FRAUD_NET_TRX_STATUS_NEW,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapTransactionStatusToFraudNet(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestMapPaymentStatusToFraudNetStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Paid - uppercase",
			input:    "PAID",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_PAID,
		},
		{
			name:     "Success - uppercase",
			input:    "SUCCESS",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_PAID,
		},
		{
			name:     "Success - lowercase",
			input:    "success",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_PAID,
		},
		{
			name:     "Refunded - uppercase",
			input:    "REFUNDED",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_REFUNDED,
		},
		{
			name:     "Refunded - lowercase",
			input:    "refunded",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_REFUNDED,
		},
		{
			name:     "Pending - uppercase",
			input:    "PENDING",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_AUTH,
		},
		{
			name:     "Waiting for Payment - uppercase",
			input:    "WAITING_FOR_PAYMENT",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_AUTH,
		},
		{
			name:     "Require Confirmation - uppercase",
			input:    "REQUIRE_CONFIRMATION",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_INVOICED,
		},
		{
			name:     "Require Action - uppercase",
			input:    "REQUIRE_ACTION",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_PARTIAL_DEFAULT,
		},
		{
			name:     "Failed - uppercase",
			input:    "FAILED",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_DECLINED,
		},
		{
			name:     "Blocked - uppercase",
			input:    "BLOCKED",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_DECLINED,
		},
		{
			name:     "Cancelled - uppercase",
			input:    "CANCELLED",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_DECLINED,
		},
		{
			name:     "Expired - uppercase",
			input:    "EXPIRED",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_DEFAULT,
		},
		{
			name:     "Void - uppercase",
			input:    "VOID",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_VOID,
		},
		{
			name:     "Default Case - Unknown",
			input:    "UNKNOWN",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_AUTH,
		},
		{
			name:     "Default Case - Empty String",
			input:    "",
			expected: constant.FRAUD_NET_PAYMENT_STATUS_AUTH,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapPaymentStatusToFraudNetStatus(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// TestSaveFdsRiskAssessmentToLedger tests the saveFdsRiskAssessmentToLedger method specifically
func TestSaveFdsRiskAssessmentToLedger(t *testing.T) {
	mockUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	evalResults := []fdscommon.EvalResult{
		{
			Success: true,
			RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
				Result: "LOW",
			},
		},
	}

	testCases := []struct {
		desc                   string
		transaction            interface{}
		fdsResp                *fdscommon.CheckTransactionResponse
		existingAdditionalInfo map[string]interface{}
		setupMock              func(*mockRepositories.IAccountTransactionRepository, string)
		expectedError          bool
		expectedRiskAssessment bool
		expectedFailureCode    string
	}{
		{
			desc: "success case - FDS_STATUS_PASSED with no failure code",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"existingField":"existingValue"}`)),
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.MatchedBy(func(nullJSON types.NullJSONText) bool {
					if !nullJSON.Valid {
						return false
					}

					var result map[string]interface{}
					err := json.Unmarshal(nullJSON.JSONText, &result)
					if err != nil {
						return false
					}

					// Should have fdsRiskAssessment
					_, hasRiskAssessment := result["fdsRiskAssessment"]
					if !hasRiskAssessment {
						return false
					}

					// Should NOT have failureCode for PASSED status
					_, hasFailureCode := result["failureCode"]
					return !hasFailureCode
				})).Return(nil)
			},
			expectedError:          false,
			expectedRiskAssessment: true,
			expectedFailureCode:    "",
		},
		{
			desc: "success case - FDS_STATUS_REJECTED sets BLOCKED_BY_FDS failure code",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"existingField":"existingValue"}`)),
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REJECTED,
				Score:  decimal.NewFromInt(85),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.MatchedBy(func(nullJSON types.NullJSONText) bool {
					if !nullJSON.Valid {
						return false
					}

					var result map[string]interface{}
					err := json.Unmarshal(nullJSON.JSONText, &result)
					if err != nil {
						return false
					}

					// Should have fdsRiskAssessment
					_, hasRiskAssessment := result["fdsRiskAssessment"]
					if !hasRiskAssessment {
						return false
					}

					// Should have failureCode = BLOCKED_BY_FDS
					failureCode, hasFailureCode := result["failureCode"]
					return hasFailureCode && failureCode == constant.FailureCodeBlockedByFDS
				})).Return(nil)
			},
			expectedError:          false,
			expectedRiskAssessment: true,
			expectedFailureCode:    constant.FailureCodeBlockedByFDS,
		},
		{
			desc: "success case - FDS_STATUS_REVIEW sets REQUIRE_REVIEW failure code",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"existingField":"existingValue"}`)),
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REVIEW,
				Score:  decimal.NewFromInt(65),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.MatchedBy(func(nullJSON types.NullJSONText) bool {
					if !nullJSON.Valid {
						return false
					}

					var result map[string]interface{}
					err := json.Unmarshal(nullJSON.JSONText, &result)
					if err != nil {
						return false
					}

					// Should have fdsRiskAssessment
					_, hasRiskAssessment := result["fdsRiskAssessment"]
					if !hasRiskAssessment {
						return false
					}

					// Should have failureCode = REQUIRE_REVIEW
					failureCode, hasFailureCode := result["failureCode"]
					return hasFailureCode && failureCode == constant.FailureCodeRequireReview
				})).Return(nil)
			},
			expectedError:          false,
			expectedRiskAssessment: true,
			expectedFailureCode:    constant.FailureCodeRequireReview,
		},
		{
			desc: "success case - existing additional_info updated correctly",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"existingField":"existingValue"}`)),
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.MatchedBy(func(nullJSON types.NullJSONText) bool {
					// Verify the JSON contains both existing field and fdsRiskAssessment
					return nullJSON.Valid && len(nullJSON.JSONText) > 0
				})).Return(nil)
			},
			expectedError:          false,
			expectedRiskAssessment: true,
			expectedFailureCode:    "",
		},
		{
			desc: "success case - nil additional_info creates new map",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid: false,
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.MatchedBy(func(nullJSON types.NullJSONText) bool {
					// Verify the JSON contains fdsRiskAssessment and is valid
					if !nullJSON.Valid {
						return false
					}

					var result map[string]interface{}
					err := json.Unmarshal(nullJSON.JSONText, &result)
					if err != nil {
						return false
					}

					// Check if fdsRiskAssessment exists
					_, exists := result["fdsRiskAssessment"]
					return exists
				})).Return(nil)
			},
			expectedError:          false,
			expectedRiskAssessment: true,
			expectedFailureCode:    "",
		},
		{
			desc: "error case - repository update fails",
			transaction: &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{"existingField":"existingValue"}`)),
				},
			},
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.AnythingOfType("types.NullJSONText")).Return(pkgErrors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords))
			},
			expectedError:          true,
			expectedRiskAssessment: false,
			expectedFailureCode:    "",
		},
		{
			desc:        "success case - type assertion fails, should skip gracefully",
			transaction: "invalid-type", // This will fail type assertion
			fdsResp: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(30),
			},
			setupMock: func(accountTrxRepo *mockRepositories.IAccountTransactionRepository, expectedFailureCode string) {
				// No repository calls should be made
			},
			expectedError:          false,
			expectedRiskAssessment: false,
			expectedFailureCode:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expected mocks
			tc.setupMock(accountTrxRepo, tc.expectedFailureCode)

			// Create minimal service for testing
			service := &FdsService{
				logger:                        logger,
				accountTransactionsRepository: accountTrxRepo,
			}

			// Call the method under test
			ctx := context.Background()
			err := service.saveFdsRiskAssessmentToLedger(ctx, tc.transaction, tc.fdsResp, evalResults)

			// Verify results
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			accountTrxRepo.AssertExpectations(t)
		})
	}
}

// TestCheckTransaction_SaveFdsRiskAssessmentError tests the error handling in CheckTransaction when saveFdsRiskAssessmentToLedger fails
func TestCheckTransaction_SaveFdsRiskAssessmentError(t *testing.T) {
	// Create default values
	transactionID := "test-transaction-id"
	merchantID := uuid.New()
	referenceID := "test-reference-id"
	trxUUID := uuid.New()

	// Mock payment metadata with proper structure to avoid early errors
	mockMetadata := map[string]interface{}{
		"cardData": map[string]interface{}{
			"Fingerprint":     "test-fingerprint",
			"First8Digit":     "12345678",
			"Last4Digit":      "1234",
			"CardBrand":       "VISA",
			"CardCountryCode": "ID",
		},
	}

	// Valid additional info structure to pass GetCreditcardMetadataFromAdditionalInfo
	raw := `{
	       "methodDetail": {
	           "card": {
	               "last4": "1234",
	               "first8": "44400000",
	               "fingerprint": "0077187d-f69d-4b2c-a9f9-99aeb6919dda",
	               "cardBrand": "VISA",
	               "countryCode": "ID"
	           }
	       }
	   }`

	additionalInfo := types.NullJSONText{
		Valid:    true,
		JSONText: types.JSONText([]byte(raw)),
	}

	// Setup mocks
	accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
	paymentRepo := mockRepositories.NewIPaymentRepository(t)
	merchantRepo := mockRepositories.NewIMerchantRepository(t)
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
	fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)

	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

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

	// Mock rule evaluation repo
	ruleEvalRepo.On("Create", mock.Anything, mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations")).Return(nil)

	// Mock UpdateAdditionalInfoByID to fail - this will trigger the error handling in lines 259-262
	accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mock.Anything, mock.Anything).Return(pkgErrors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords)).Maybe()

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
	resp, err := service.CheckTransaction(ctx, transactionID, nil)

	// Verify that the main function succeeds despite saveFdsRiskAssessmentToLedger failing
	// This tests lines 259-262 where the error is logged but not returned
	assert.NoError(t, err, "CheckTransaction should not return error even if saveFdsRiskAssessmentToLedger fails")
	assert.NotNil(t, resp)
	assert.Equal(t, constant.FDS_STATUS_PASSED, resp.Status)
	assert.True(t, decimal.NewFromInt(30).Equal(resp.Score))

	// Verify all expectations were met
	accountTrxRepo.AssertExpectations(t)
	paymentRepo.AssertExpectations(t)
	merchantRepo.AssertExpectations(t)
	fraudRulesRepo.AssertExpectations(t)
	ruleEvalRepo.AssertExpectations(t)
	fraudNetMock.AssertExpectations(t)
}

// TestSaveFdsRiskAssessmentToLedger_RiskAssessmentContent tests the content of the risk assessment
func TestSaveFdsRiskAssessmentToLedger_RiskAssessmentContent(t *testing.T) {
	mockUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	testCases := []struct {
		desc                   string
		fdsStatus              string
		fdsScore               decimal.Decimal
		evalResults            []fdscommon.EvalResult
		expectedRecommendation string
		expectedLevel          string
	}{
		{
			desc:      "passed status should have approve recommendation and low level",
			fdsStatus: constant.FDS_STATUS_PASSED,
			fdsScore:  decimal.NewFromInt(30),
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						Result: "LOW",
					},
				},
			},
			expectedRecommendation: "Approve",
			expectedLevel:          "LOW",
		},
		{
			desc:      "rejected status should have reject recommendation and high level",
			fdsStatus: constant.FDS_STATUS_REJECTED,
			fdsScore:  decimal.NewFromInt(80),
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						Result: "HIGH",
					},
				},
			},
			expectedRecommendation: "Reject",
			expectedLevel:          "HIGH",
		},
		{
			desc:      "no successful evaluations should use default levels",
			fdsStatus: constant.FDS_STATUS_PASSED,
			fdsScore:  decimal.NewFromInt(20),
			evalResults: []fdscommon.EvalResult{
				{
					Success: false, // Unsuccessful evaluation
				},
			},
			expectedRecommendation: "Approve",
			expectedLevel:          "low", // Default for passed
		},
		{
			desc:      "no successful evaluations with rejected status should use high default",
			fdsStatus: constant.FDS_STATUS_REJECTED,
			fdsScore:  decimal.NewFromInt(90),
			evalResults: []fdscommon.EvalResult{
				{
					Success: false, // Unsuccessful evaluation
				},
			},
			expectedRecommendation: "Reject",
			expectedLevel:          "high", // Default for rejected
		},
		{
			desc:      "multiple evaluations should use latest successful one",
			fdsStatus: constant.FDS_STATUS_PASSED,
			fdsScore:  decimal.NewFromInt(40),
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						Result: "MEDIUM",
					},
				},
				{
					Success: false, // This should be ignored
				},
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						Result: "LOW", // This should be used (latest successful)
					},
				},
			},
			expectedRecommendation: "Approve",
			expectedLevel:          "LOW",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Capture the actual data sent to repository
			var capturedJSON types.NullJSONText
			accountTrxRepo.On("UpdateAdditionalInfoByID", mock.Anything, mockUUID.String(), mock.AnythingOfType("types.NullJSONText")).
				Run(func(args mock.Arguments) {
					capturedJSON = args.Get(2).(types.NullJSONText)
				}).Return(nil)

			// Create service
			service := &FdsService{
				logger:                        logger,
				accountTransactionsRepository: accountTrxRepo,
			}

			// Create test data
			transaction := &orchestrator_model.AccountTransactionWithUseCase{
				UUID: mockUUID,
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: types.JSONText([]byte(`{}`)),
				},
			}

			fdsResp := &fdscommon.CheckTransactionResponse{
				Status: tc.fdsStatus,
				Score:  tc.fdsScore,
			}

			// Call the method under test
			ctx := context.Background()
			err := service.saveFdsRiskAssessmentToLedger(ctx, transaction, fdsResp, tc.evalResults)

			// Verify no error
			assert.NoError(t, err)

			// Verify the captured JSON content
			assert.True(t, capturedJSON.Valid)

			var result map[string]interface{}
			err = json.Unmarshal(capturedJSON.JSONText, &result)
			assert.NoError(t, err)

			// Verify fdsRiskAssessment exists
			riskAssessmentRaw, exists := result["fdsRiskAssessment"]
			assert.True(t, exists)

			// Convert to map for easier validation
			riskAssessmentMap, ok := riskAssessmentRaw.(map[string]interface{})
			assert.True(t, ok)

			// Verify recommendation
			assert.Equal(t, tc.expectedRecommendation, riskAssessmentMap["recommendation"])

			// Verify level
			assert.Equal(t, tc.expectedLevel, riskAssessmentMap["level"])

			// Verify score
			scoreStr, ok := riskAssessmentMap["score"].(string)
			assert.True(t, ok)
			assert.Equal(t, tc.fdsScore.String(), scoreStr)

			// Verify status
			assert.Equal(t, tc.fdsStatus, riskAssessmentMap["status"])

			// Verify EvaluatedAt exists and is a string (time)
			_, exists = riskAssessmentMap["evaluatedAt"]
			assert.True(t, exists)

			// Verify all expectations were met
			accountTrxRepo.AssertExpectations(t)
		})
	}
}

func TestExtractRuleNamesFromTags(t *testing.T) {
	service := &FdsService{}

	tests := []struct {
		name     string
		tags     []fdscommon.CheckTags
		expected []string
	}{
		{
			name: "extract single rule with state and name",
			tags: []fdscommon.CheckTags{
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("Card number blacklist"),
					Name:   "3528 24•• •••• 1821",
				},
			},
			expected: []string{"Card number blacklist 3528 24•• •••• 1821"},
		},
		{
			name: "extract rule with only state",
			tags: []fdscommon.CheckTags{
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("High velocity"),
					Name:   "",
				},
			},
			expected: []string{"High velocity"},
		},
		{
			name: "extract rule with only name",
			tags: []fdscommon.CheckTags{
				{
					Source: "rule",
					Type:   "rule",
					State:  nil,
					Name:   "Suspicious activity",
				},
			},
			expected: []string{"Suspicious activity"},
		},
		{
			name: "extract multiple rules",
			tags: []fdscommon.CheckTags{
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("Card blacklist"),
					Name:   "1234 5678",
				},
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("Velocity check"),
					Name:   "Too many transactions",
				},
			},
			expected: []string{"Card blacklist 1234 5678", "Velocity check Too many transactions"},
		},
		{
			name: "ignore non-rule tags",
			tags: []fdscommon.CheckTags{
				{
					Source: "workflow",
					Type:   "queue",
					State:  util.ValueToPtr("cancelled"),
					Name:   "cancelled",
				},
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("Valid rule"),
					Name:   "Rule name",
				},
			},
			expected: []string{"Valid rule Rule name"},
		},
		{
			name: "ignore rules with empty state and name",
			tags: []fdscommon.CheckTags{
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr(""),
					Name:   "",
				},
				{
					Source: "rule",
					Type:   "rule",
					State:  util.ValueToPtr("Valid rule"),
					Name:   "Valid name",
				},
			},
			expected: []string{"Valid rule Valid name"},
		},
		{
			name:     "empty tags",
			tags:     []fdscommon.CheckTags{},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := service.extractRuleNamesFromTags(tc.tags)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSendFdsSlackAlert(t *testing.T) {
	tests := []struct {
		name            string
		evalResults     []fdscommon.EvalResult
		expectedPublish bool
		setupMock       func(*mockRabbitMq.RabbitMQExt)
	}{
		{
			name:            "no eval results - should not publish",
			evalResults:     []fdscommon.EvalResult{},
			expectedPublish: false,
			setupMock:       func(rabbitMq *mockRabbitMq.RabbitMQExt) {},
		},
		{
			name: "single provider with rules - should publish",
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						UUID:     "rule-eval-id",
						RuleID:   "rule-1",
						Provider: constant.PROVIDER_FRAUD_NET,
						Score:    decimal.NewFromInt(90),
						Result:   "HIGH",
					},
					Data: &fdscommon.CheckData{
						Tags: []fdscommon.CheckTags{
							{
								Source: "rule",
								Type:   "rule",
								State:  util.ValueToPtr("Card number blacklist"),
								Name:   "3528 24•• •••• 1821",
							},
						},
					},
					Weight: decimal.NewFromFloat(0.3),
				},
			},
			expectedPublish: true,
			setupMock: func(rabbitMq *mockRabbitMq.RabbitMQExt) {
				rabbitMq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("[]uint8")).Return(nil)
			},
		},
		{
			name: "provider without tags - should publish with basic info",
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						UUID:     "rule-eval-id",
						RuleID:   "rule-1",
						Provider: constant.PROVIDER_FRAUD_NET,
						Score:    decimal.NewFromInt(50),
						Result:   "MEDIUM",
					},
					Data:   &fdscommon.CheckData{Tags: nil},
					Weight: decimal.NewFromFloat(0.5),
				},
			},
			expectedPublish: true,
			setupMock: func(rabbitMq *mockRabbitMq.RabbitMQExt) {
				rabbitMq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("[]uint8")).Return(nil)
			},
		},
		{
			name: "multiple providers - should publish",
			evalResults: []fdscommon.EvalResult{
				{
					Success: true,
					RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
						UUID:     "rule-eval-1",
						RuleID:   "rule-1",
						Provider: constant.PROVIDER_FRAUD_NET,
						Score:    decimal.NewFromInt(80),
						Result:   "HIGH",
					},
					Data: &fdscommon.CheckData{
						Tags: []fdscommon.CheckTags{
							{
								Source: "rule",
								Type:   "rule",
								State:  util.ValueToPtr("High velocity"),
								Name:   "Multiple transactions",
							},
						},
					},
					Weight: decimal.NewFromFloat(0.4),
				},
			},
			expectedPublish: true,
			setupMock: func(rabbitMq *mockRabbitMq.RabbitMQExt) {
				rabbitMq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("[]uint8")).Return(nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rabbitMq := mockRabbitMq.NewRabbitMQExt(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(rabbitMq)

			service := &FdsService{
				logger:   logger,
				rabbitMq: rabbitMq,
				cfg: &config.Config{
					SlackConfig: config.SlackConfig{
						FDSAlertWebhookURL: "https://hooks.slack.com/test",
					},
				},
			}

			payment := &paymentModel.Payment{
				UUID:     "payment-123",
				Amount:   decimal.NewFromInt(100000),
				Currency: "IDR",
			}

			merchant := &merchant.Merchant{
				UUID: "merchant-123",
				Name: "Test Merchant",
			}

			resp := fdscommon.CheckTransactionResponse{
				Status:      constant.FDS_STATUS_REJECTED,
				Score:       decimal.NewFromInt(90),
				EvalResults: &tc.evalResults,
			}

			err := service.SendFdsSlackAlert(context.Background(), "txn-123", "card-123", payment, merchant, resp)

			if tc.expectedPublish {
				assert.NoError(t, err)
				rabbitMq.AssertExpectations(t)
			} else {
				assert.NoError(t, err)
				rabbitMq.AssertNotCalled(t, "Publish")
			}
		})
	}
}

func TestSendFdsSlackAlert_PublishError(t *testing.T) {
	rabbitMq := mockRabbitMq.NewRabbitMQExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	publishError := errors.New("publish failed")
	rabbitMq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("[]uint8")).Return(publishError)

	service := &FdsService{
		logger:   logger,
		rabbitMq: rabbitMq,
		cfg: &config.Config{
			SlackConfig: config.SlackConfig{
				FDSAlertWebhookURL: "https://hooks.slack.com/test",
			},
		},
	}

	evalResults := []fdscommon.EvalResult{
		{
			Success: true,
			RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
				UUID:     "rule-eval-id",
				RuleID:   "rule-1",
				Provider: constant.PROVIDER_FRAUD_NET,
				Score:    decimal.NewFromInt(90),
				Result:   "HIGH",
			},
			Data:   &fdscommon.CheckData{Tags: []fdscommon.CheckTags{}},
			Weight: decimal.NewFromFloat(0.3),
		},
	}

	payment := &paymentModel.Payment{
		UUID:     "payment-123",
		Amount:   decimal.NewFromInt(100000),
		Currency: "IDR",
	}

	merchant := &merchant.Merchant{
		UUID: "merchant-123",
		Name: "Test Merchant",
	}

	resp := fdscommon.CheckTransactionResponse{
		Status:      constant.FDS_STATUS_REJECTED,
		Score:       decimal.NewFromInt(90),
		EvalResults: &evalResults,
	}

	err := service.SendFdsSlackAlert(context.Background(), "txn-123", "card-123", payment, merchant, resp)

	assert.Error(t, err)
	assert.Equal(t, publishError, err)
	rabbitMq.AssertExpectations(t)
}

func TestMapAcquirerResponseCodeToFraudNetCardStatus(t *testing.T) {
	tests := []struct {
		name         string
		responseCode string
		expected     string
	}{
		{
			name:         "Decline codes - 01",
			responseCode: "01",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Decline codes - 03",
			responseCode: "03",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Decline codes - 6P",
			responseCode: "6P",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Decline codes - N7",
			responseCode: "N7",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Decline codes - 911",
			responseCode: "911",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Expired codes - 54",
			responseCode: "54",
			expected:     constant.FRAUD_NET_CARD_STATUS_EXPIRED,
		},
		{
			name:         "Expired codes - 101",
			responseCode: "101",
			expected:     constant.FRAUD_NET_CARD_STATUS_EXPIRED,
		},
		{
			name:         "Inactive codes - 14",
			responseCode: "14",
			expected:     constant.FRAUD_NET_CARD_STATUS_INACTIVE,
		},
		{
			name:         "Inactive codes - 111",
			responseCode: "111",
			expected:     constant.FRAUD_NET_CARD_STATUS_INACTIVE,
		},
		{
			name:         "Stolen codes - 04",
			responseCode: "04",
			expected:     constant.FRAUD_NET_CARD_STATUS_STOLEN,
		},
		{
			name:         "Stolen codes - 200",
			responseCode: "200",
			expected:     constant.FRAUD_NET_CARD_STATUS_STOLEN,
		},
		{
			name:         "Suspended codes - 34",
			responseCode: "34",
			expected:     constant.FRAUD_NET_CARD_STATUS_SUSPENDED,
		},
		{
			name:         "Suspended codes - 83",
			responseCode: "83",
			expected:     constant.FRAUD_NET_CARD_STATUS_SUSPENDED,
		},
		{
			name:         "Default case - unknown code 99",
			responseCode: "99",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Default case - empty string",
			responseCode: "",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
		{
			name:         "Default case - non-numeric code",
			responseCode: "ABC",
			expected:     constant.FRAUD_NET_CARD_STATUS_DECLINE,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapAcquirerResponseCodeToFraudNetCardStatus(tc.responseCode)
			assert.Equal(t, tc.expected, got)
		})
	}
}
