package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetCustomLimitConfig(t *testing.T) {
	merchantID := "test-merchant-id"
	bankCode := "008"
	accountNo := "1234567890"

	testCases := []struct {
		name      string
		request   merchant.BeneficiaryLimitConfigRequest
		mockSetup func(
			mockRepo *mocks.IMerchantRepository,
			mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
			mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
		)
		wantErr bool
	}{
		{
			name: "SUCCESS: Set merchant policy rule (empty beneficiary bank code and account no)",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				existingMerchant := &merchant.Merchant{
					UUID: merchantID,
					Name: "Test Merchant",
					Metadata: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"kymNotes":"test notes"}`),
					},
				}
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(existingMerchant, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update merchant policy rule to nil (remove limit)",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:                 merchantID,
				BeneficiaryBankCode:        "",
				BeneficiaryAccountNo:       "",
				BeneficiaryPayoutLimitRule: nil,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				existingMerchant := &merchant.Merchant{
					UUID: merchantID,
					Name: "Test Merchant",
					Metadata: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"kymNotes":"test notes","beneficiaryPayoutLimitRule":{"velocity":10,"timeframe":"DAILY","amountThreshold":5000000}}`),
					},
				}
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(existingMerchant, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Merchant policy - FindMerchantByID returns error",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant policy - Merchant not found",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant policy - Update merchant fails",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				existingMerchant := &merchant.Merchant{
					UUID: merchantID,
					Name: "Test Merchant",
					Metadata: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{"kymNotes":"test notes"}`),
					},
				}
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(existingMerchant, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Set beneficiary account custom limit",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
					},
				}
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(beneficiaryAccount, nil)
				mockBeneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Remove beneficiary account custom limit (set to nil)",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:                 merchantID,
				BeneficiaryBankCode:        bankCode,
				BeneficiaryAccountNo:       accountNo,
				BeneficiaryPayoutLimitRule: nil,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
						BeneficiaryPayoutLimitRule: &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
							Velocity:        15,
							Timeframe:       "DAILY",
							AmountThreshold: 10000000,
						},
					},
				}
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(beneficiaryAccount, nil)
				mockBeneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Beneficiary account update returns ErrNoRowsAffected (treated as success)",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
					},
				}
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(beneficiaryAccount, nil)
				mockBeneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Beneficiary account - FindByBankCodeAndAccountNo returns error",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Beneficiary account - Beneficiary not found",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Beneficiary account - Update fails",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
					},
				}
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(beneficiaryAccount, nil)
				mockBeneficiaryRepo.On("Update", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant policy - Invalid merchant metadata JSON",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				// Merchant with invalid JSON metadata
				existingMerchant := &merchant.Merchant{
					UUID: merchantID,
					Name: "Test Merchant",
					Metadata: types.NullJSONText{
						Valid:    true,
						JSONText: []byte(`{invalid json}`),
					},
				}
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(existingMerchant, nil)
				mockRepo.On("Update", mock.Anything, existingMerchant).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Beneficiary account path - Verify request parameters",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  bankCode,
				BeneficiaryAccountNo: accountNo,
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        15,
					Timeframe:       "DAILY",
					AmountThreshold: 10000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
					},
				}

				// Verify the request parameters passed to FindByBankCodeAndAccountNo
				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.MatchedBy(func(req *beneficiaryModel.CheckAccountRequest) bool {
					return req.MerchantID == merchantID &&
						req.BeneficiaryBankCode == bankCode &&
						req.BeneficiaryAccountNo == accountNo
				})).Return(beneficiaryAccount, nil)

				mockBeneficiaryRepo.On("Update", mock.Anything, mock.MatchedBy(func(ba *beneficiaryModel.BeneficiaryAccount) bool {
					// Verify that the metadata was properly set
					return ba.Metadata.Valid && ba.MetadataObj.BeneficiaryPayoutLimitRule != nil &&
						ba.MetadataObj.BeneficiaryPayoutLimitRule.Velocity == 15 &&
						ba.MetadataObj.BeneficiaryPayoutLimitRule.AmountThreshold == 10000000
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Beneficiary account path - Verify nil rule removes limit",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:                 merchantID,
				BeneficiaryBankCode:        bankCode,
				BeneficiaryAccountNo:       accountNo,
				BeneficiaryPayoutLimitRule: nil,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				beneficiaryAccount := &beneficiaryModel.Account{
					UUID:                   "beneficiary-uuid",
					MerchantID:             merchantID,
					BeneficiaryAccountNo:   accountNo,
					BeneficiaryAccountName: "Test Beneficiary",
					BeneficiaryBankCode:    bankCode,
					BeneficiaryBankName:    "Test Bank",
					MetadataObj: beneficiaryModel.Metadata{
						IsXb:          false,
						IsOverbooking: false,
						MaxAmount:     decimal.NewFromInt(0),
						BeneficiaryPayoutLimitRule: &disbursementModel.BeneficiaryPayoutLimitRuleConfig{
							Velocity:        15,
							Timeframe:       "DAILY",
							AmountThreshold: 10000000,
						},
					},
				}

				mockBeneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(beneficiaryAccount, nil)

				// Verify that the limit rule is set to nil
				mockBeneficiaryRepo.On("Update", mock.Anything, mock.MatchedBy(func(ba *beneficiaryModel.BeneficiaryAccount) bool {
					return ba.Metadata.Valid && ba.MetadataObj.BeneficiaryPayoutLimitRule == nil
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Check wrapping of database errors",
			request: merchant.BeneficiaryLimitConfigRequest{
				MerchantID:           merchantID,
				BeneficiaryBankCode:  "",
				BeneficiaryAccountNo: "",
				BeneficiaryPayoutLimitRule: &merchant.BeneficiaryLimitConfigRuleRequest{
					Velocity:        10,
					Timeframe:       "DAILY",
					AmountThreshold: 5000000,
				},
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockBeneficiaryRepo *mocks.IBeneficiaryAccountRepository,
				mockBeneficiaryAccountSvc *serviceMocks.IBeneficiaryAccountService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("database connection error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewIMerchantRepository(t)
			mockBeneficiaryRepo := mocks.NewIBeneficiaryAccountRepository(t)
			mockBeneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockRepo, mockBeneficiaryRepo, mockBeneficiaryAccountSvc)

			svc := New(
				mockRepo,
				mockLogger,
				nil,
				nil,
				nil,
				nil,
				WithBeneficiaryAccountRepo(mockBeneficiaryRepo),
				WithBeneficiaryAccountService(mockBeneficiaryAccountSvc),
			)

			err := svc.SetCustomLimitConfig(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				// Special case: Check if error wrapping is tested
				if tc.name == "ERROR: Check wrapping of database errors" {
					assert.ErrorContains(t, err, response.HttpErrDatabase)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockBeneficiaryRepo.AssertExpectations(t)
			mockBeneficiaryAccountSvc.AssertExpectations(t)
		})
	}
}
