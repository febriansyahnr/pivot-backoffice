package account_repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindMerchantAccountByName(t *testing.T) {
	accountType := constant.TypePayment
	merchantID := uuid.New()
	accountResult := &account_model.Account{
		UUID:        uuid.New(),
		ReferenceID: merchantID,
		Name:        accountType,
		EODBalance:  12500,
		Currency:    "IDR",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *account_model.Account
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Merchant Balance By Name",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					accountPtr := args.Get(1).(*account_model.Account)
					*accountPtr = *accountResult
				})
			},
			expected: accountResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Account Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Find Merchant Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "accounts")
			transaction, err := repo.FindMerchantAccountByName(ctx, merchantID, accountType)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetByUUID(t *testing.T) {
	merchantID := uuid.New()
	accountResult := &account_model.Account{
		UUID:        uuid.New(),
		ReferenceID: merchantID,
		Name:        constant.TypePayment,
		EODBalance:  12500,
		Currency:    "IDR",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *account_model.Account
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get by UUID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					balancePtr := args.Get(1).(*account_model.Account)
					*balancePtr = *accountResult
				})
			},
			expected: accountResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Balance Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Get By UUID Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.PtrAccountMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "accounts")
			transaction, err := repo.GetByUUID(ctx, uuid.New())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetByIDs(t *testing.T) {
	merchantID := uuid.New()
	accountResult := []*account_model.Account{
		&account_model.Account{
			UUID:        uuid.New(),
			ReferenceID: merchantID,
			Name:        constant.TypePayment,
			EODBalance:  12500,
			Currency:    "IDR",
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []*account_model.Account
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get by IDs",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					balancePtr := args.Get(1).(*[]*account_model.Account)
					*balancePtr = accountResult
				})
			},
			expected: accountResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Balance Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Get By UUID Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "accounts")
			transaction, err := repo.GetByIDs(ctx, []string{uuid.New().String()})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindAll(t *testing.T) {
	merchantID := uuid.New()
	accountResult := []*account_model.Account{
		{
			UUID:        uuid.New(),
			ReferenceID: merchantID,
			Name:        constant.TypePayment,
			EODBalance:  12500,
			Currency:    "IDR",
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []*account_model.Account
		wantErr   bool
	}{
		{
			name: "SUCCESS: FindAll",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*account_model.Account"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					balancePtr := args.Get(1).(*[]*account_model.Account)
					*balancePtr = accountResult
				})
			},
			expected: accountResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Find All Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*account_model.Account"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "accounts")
			transaction, err := repo.FindAll(ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetEntityAccounts(t *testing.T) {
	firstEntity := uuid.New()
	secondEntity := uuid.New()
	accountList := []*account_model.Account{
		{
			UUID:        uuid.New(),
			ReferenceID: firstEntity,
		},
		{
			UUID:        uuid.New(),
			ReferenceID: secondEntity,
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  map[uuid.UUID]*account_model.Account
		wantErr   bool
		entityIDs []uuid.UUID
	}{
		{
			name: "SUCCESS: GetEntityAccounts",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*account_model.Account"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(string(constant.StringMockType())),
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(nil).Run(func(args mock.Arguments) {
					accounts := args.Get(1).(*[]*account_model.Account)
					*accounts = accountList
				})

				mysqlMock.On("Rebind", constant.StringMockType()).Return(mock.Anything)
			},
			expected: map[uuid.UUID]*account_model.Account{
				firstEntity:  accountList[0],
				secondEntity: accountList[1],
			},
			wantErr:   false,
			entityIDs: []uuid.UUID{firstEntity, secondEntity},
		},
		{
			name: "ERROR: Fail sql IN",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
			},
			expected: map[uuid.UUID]*account_model.Account{
				firstEntity:  accountList[0],
				secondEntity: accountList[1],
			},
			wantErr:   true,
			entityIDs: []uuid.UUID{},
		},
		{
			name: "ERROR: Select",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*account_model.Account"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(string(constant.StringMockType())),
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(errors.New("errors"))

				mysqlMock.On("Rebind", constant.StringMockType()).Return(mock.Anything)
			},
			expected: map[uuid.UUID]*account_model.Account{
				firstEntity:  accountList[0],
				secondEntity: accountList[1],
			},
			wantErr:   true,
			entityIDs: []uuid.UUID{firstEntity, secondEntity},
		},
		{
			name: "ERROR: Select No Rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*account_model.Account"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(string(constant.StringMockType())),
					mock.AnythingOfType(string(constant.StringMockType())),
				).Return(sql.ErrNoRows)

				mysqlMock.On("Rebind", constant.StringMockType()).Return(mock.Anything)
			},
			expected:  nil,
			wantErr:   false,
			entityIDs: []uuid.UUID{firstEntity, secondEntity},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), constant.CtxSQLTableNameKey, "accounts")
			accounts, err := repo.GetEntityAccounts(ctx, tc.entityIDs, constant.UserTypeMerchant, constant.ReferenceDisbursement)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, accounts)
			}
			mockMysql.AssertExpectations(t)

		})
	}

}
