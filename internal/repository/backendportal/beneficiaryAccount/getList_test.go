package beneficiaryAccountRepository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeneficiaryAccountRepository_GetList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(errors.New("no rows data"))
			},
			filter: &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{
				MerchantID:             uuid.NewString(),
				BeneficiaryAccountNo:   "beneficiary-account-no",
				BeneficiaryAccountName: "beneficiary-account-name",
				StartCreatedAt:         &util.TimeNow,
				EndCreatedAt:           &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter xb",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(errors.New("no rows data"))
			},
			filter: &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{
				MerchantID:             uuid.NewString(),
				BeneficiaryAccountNo:   "beneficiary-account-no",
				BeneficiaryAccountName: "beneficiary-account-name",
				StartCreatedAt:         &util.TimeNow,
				EndCreatedAt:           &util.TimeNow,
				IsXb:                   true,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
				).Return(errors.New("some-error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)

			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get List with valid metadata that gets unmarshaled",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					// Get the pointer to the data slice
					accounts := args.Get(1).(*[]*beneficiaryAccountModel.BeneficiaryAccount)

					// Create test data with valid metadata JSON
					metadataJSON := []byte(`{"isXb": true, "maxAmount": "1000.00", "isOverbooking": false}`)

					// Fill the slice with test data
					*accounts = []*beneficiaryAccountModel.BeneficiaryAccount{
						{
							UUID:                   "test-uuid-1",
							BeneficiaryAccountNo:   "1234567890",
							BeneficiaryAccountName: "Test Account",
							BeneficiaryBankCode:    "001",
							BeneficiaryBankName:    "Test Bank",
							Metadata: types.NullJSONText{
								Valid:    true,
								JSONText: metadataJSON,
							},
						},
						{
							UUID:                   "test-uuid-2",
							BeneficiaryAccountNo:   "0987654321",
							BeneficiaryAccountName: "Test Account 2",
							BeneficiaryBankCode:    "002",
							BeneficiaryBankName:    "Test Bank 2",
							// This one has Metadata.Valid = false
							Metadata: types.NullJSONText{
								Valid: false,
							},
						},
					}
				}).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					totalItems := args.Get(1).(*int64)
					*totalItems = 2
				}).Return(nil)
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			response, err := repo.GetList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Additional assertions for the metadata unmarshaling test case
				if tc.name == "SUCCESS: Get List with valid metadata that gets unmarshaled" {
					accounts := response.Data.([]*beneficiaryAccountModel.BeneficiaryAccount)

					// Check that we have 2 accounts
					assert.Equal(t, 2, len(accounts))

					// Verify the first account's metadata was correctly unmarshaled
					assert.True(t, accounts[0].MetadataObj.IsXb)
					assert.False(t, accounts[0].MetadataObj.IsOverbooking)

					// Verify the second account's metadata is empty (not unmarshaled)
					assert.False(t, accounts[1].MetadataObj.IsXb)
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestBeneficiaryAccountRepository_GetListOfDerived(t *testing.T) {
	derivedMerchantID := uuid.NewString()
	
	testCase := []struct {
		name      string
		ctx       context.Context
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List of derived without any filter",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						// Verify query contains JOIN with disbursements and GROUP BY
						return strings.Contains(query, "JOIN disbursements d") && 
							   strings.Contains(query, "GROUP BY ba.uuid")
					}),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						// Verify count query contains JOIN with disbursements
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Return(nil)
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List of derived without any filter and total items is zero",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d") && 
							   strings.Contains(query, "GROUP BY ba.uuid")
					}),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Return(errors.New("no rows data"))
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List of derived with MerchantID filter uses derivedMerchantID",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d") && 
							   strings.Contains(query, "ba.merchant_id = ?") &&
							   strings.Contains(query, "d.merchant_id = ?")
					}),
					derivedMerchantID, // First occurrence for ba.merchant_id
					derivedMerchantID, // Second occurrence for d.merchant_id
					mock.AnythingOfType("string"), // beneficiary_account_no
					mock.AnythingOfType("string"), // beneficiary_account_name
					mock.AnythingOfType(constant.MockTypeTimeReference), // start created_at
					mock.AnythingOfType(constant.MockTypeTimeReference), // end created_at
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d") &&
							   strings.Contains(query, "ba.merchant_id = ?") &&
							   strings.Contains(query, "d.merchant_id = ?")
					}),
					derivedMerchantID, // First occurrence for ba.merchant_id
					derivedMerchantID, // Second occurrence for d.merchant_id
					mock.AnythingOfType("string"), // beneficiary_account_no
					mock.AnythingOfType("string"), // beneficiary_account_name
					mock.AnythingOfType(constant.MockTypeTimeReference), // start created_at
					mock.AnythingOfType(constant.MockTypeTimeReference), // end created_at
				).Return(errors.New("no rows data"))
			},
			filter: &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{
				MerchantID:             "original-merchant-id", // This should be ignored, derivedMerchantID used instead
				BeneficiaryAccountNo:   "beneficiary-account-no",
				BeneficiaryAccountName: "beneficiary-account-name",
				StartCreatedAt:         &util.TimeNow,
				EndCreatedAt:           &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List of derived with IsXb filter",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d") && 
							   strings.Contains(query, "ba.metadata->>'$.isXb' = 'true'")
					}),
					derivedMerchantID,
					derivedMerchantID,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d") &&
							   strings.Contains(query, "ba.metadata->>'$.isXb' = 'true'")
					}),
					derivedMerchantID,
					derivedMerchantID,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
				).Return(errors.New("no rows data"))
			},
			filter: &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{
				MerchantID:             "original-merchant-id",
				BeneficiaryAccountNo:   "beneficiary-account-no",
				BeneficiaryAccountName: "beneficiary-account-name",
				StartCreatedAt:         &util.TimeNow,
				EndCreatedAt:           &util.TimeNow,
				IsXb:                   true,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List of derived on SelectContext error",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Return(errors.New("select-error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Return(nil)
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get List of derived with valid metadata that gets unmarshaled",
			ctx:  context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*beneficiaryAccountModel.BeneficiaryAccount"),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Run(func(args mock.Arguments) {
					accounts := args.Get(1).(*[]*beneficiaryAccountModel.BeneficiaryAccount)
					metadataJSON := []byte(`{"isXb": true, "maxAmount": "1000.00", "isOverbooking": false}`)

					*accounts = []*beneficiaryAccountModel.BeneficiaryAccount{
						{
							UUID:                   "derived-uuid-1",
							BeneficiaryAccountNo:   "1234567890",
							BeneficiaryAccountName: "Derived Account",
							BeneficiaryBankCode:    "001",
							BeneficiaryBankName:    "Test Bank",
							Metadata: types.NullJSONText{
								Valid:    true,
								JSONText: metadataJSON,
							},
						},
						{
							UUID:                   "derived-uuid-2",
							BeneficiaryAccountNo:   "0987654321",
							BeneficiaryAccountName: "Derived Account 2",
							BeneficiaryBankCode:    "002",
							BeneficiaryBankName:    "Test Bank 2",
							Metadata: types.NullJSONText{
								Valid: false,
							},
						},
					}
				}).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "JOIN disbursements d")
					}),
				).Run(func(args mock.Arguments) {
					totalItems := args.Get(1).(*int64)
					*totalItems = 2
				}).Return(nil)
			},
			filter:  &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			response, err := repo.GetListOfDerived(tc.ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Additional assertions for the metadata unmarshaling test case
				if tc.name == "SUCCESS: Get List of derived with valid metadata that gets unmarshaled" {
					accounts := response.Data.([]*beneficiaryAccountModel.BeneficiaryAccount)

					// Check that we have 2 accounts
					assert.Equal(t, 2, len(accounts))

					// Verify the first account's metadata was correctly unmarshaled
					assert.True(t, accounts[0].MetadataObj.IsXb)
					assert.False(t, accounts[0].MetadataObj.IsOverbooking)

					// Verify the second account's metadata is empty (not unmarshaled)
					assert.False(t, accounts[1].MetadataObj.IsXb)
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
