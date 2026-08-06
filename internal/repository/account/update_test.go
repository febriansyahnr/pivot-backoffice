package account_repository

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateAccount(t *testing.T) {

	log := logMock.NewILogger(t)

	account := &accountModel.Account{
		UUID:                       uuid.New(),
		EODBalance:                 100000.00,
		LastUpdateBalanceAt:        time.Now().UTC().Add(-time.Minute),
		PendingTransactionCutoffAt: util.ValueToPtr(time.Now().UTC()),
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext", mock.Anything, mock.Anything, account.EODBalance, account.LastUpdateBalanceAt, account.PendingTransactionCutoffAt, account.UUID,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext", mock.Anything, mock.Anything, account.EODBalance, account.LastUpdateBalanceAt, account.PendingTransactionCutoffAt, account.UUID,
				).Return(false, constant.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "error when updating account", mock.Anything).Once().Return()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, log)
			err := repo.UpdateAccount(t.Context(), account)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdateHoldedBalance(t *testing.T) {
	log := logMock.NewILogger(t)

	account := &accountModel.Account{
		UUID:          uuid.New(),
		HoldedBalance: 50000.00,
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update holded balance",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext", mock.Anything, mock.Anything, account.HoldedBalance, account.UUID,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext", mock.Anything, mock.Anything, account.HoldedBalance, account.UUID,
				).Return(false, constant.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "error when updating holded balance", mock.Anything).Once().Return()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, log)
			err := repo.UpdateHoldedBalance(t.Context(), account)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}
