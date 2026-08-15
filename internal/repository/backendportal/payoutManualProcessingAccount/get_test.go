package payoutManualProcessingAccount_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payoutManualProcessingAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/payoutManualProcessingAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	accounts := []*payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		{
			UUID:          "uuid-123",
			MerchantID:    "merchant-123",
			BankCode:      "BCA",
			AccountNumber: "1234567890",
			Status:        constant.StatusActive,
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		query     *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery
		wantErr   bool
	}{
		{
			name: "SUCCESS: List accounts",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 1
					}).
					Return(nil)
				mysqlMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*payoutManualProcessingAccountModel.PayoutManualProcessingAccount)
						*resultPtr = accounts
					}).
					Return(nil)
			},
			query: &payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery{
				Status:   constant.StatusActive,
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Count query fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("count error"))
			},
			query: &payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery{
				Status:   constant.StatusActive,
				Page:     1,
				PageSize: 10,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Select query fails",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 1
					}).
					Return(nil)
				mysqlMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("select error"))
			},
			query: &payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery{
				Status:   constant.StatusActive,
				Page:     1,
				PageSize: 10,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payout_manual_processing_accounts")
			result, _, err := repo.List(ctx, tc.query)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(accounts), len(result))
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestIsManualProcessingAccount(t *testing.T) {
	account := &payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		UUID:          "uuid-123",
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		Status:        constant.StatusActive,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Account exists",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					accountPtr := args.Get(1).(*payoutManualProcessingAccountModel.PayoutManualProcessingAccount)
					*accountPtr = *account
				})
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Account not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payout_manual_processing_accounts")
			result, err := repo.IsManualProcessingAccount(ctx, "merchant-123", "BCA", "1234567890")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetByUUID(t *testing.T) {
	account := &payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		UUID:          "uuid-123",
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		Status:        constant.StatusActive,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		uuid      string
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get account by UUID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					accountPtr := args.Get(1).(*payoutManualProcessingAccountModel.PayoutManualProcessingAccount)
					*accountPtr = *account
				})
			},
			uuid:    "uuid-123",
			wantNil: false,
			wantErr: false,
		},
		{
			name: "SUCCESS: Account not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			uuid:    "missing-uuid",
			wantNil: true,
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			uuid:    "uuid-123",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payout_manual_processing_accounts")
			result, err := repo.GetByUUID(ctx, tc.uuid)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tc.wantNil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, account.UUID, result.UUID)
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
