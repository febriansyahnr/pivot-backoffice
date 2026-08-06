package adjustment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func TestReleaseHoldedMerchantBalance(t *testing.T) {
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	accountRepoMock := repoMocks.NewIAccountRepository(t)
	loggerMock := loggerMocks.NewILogger(t)

	loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	svc := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithAccountRepository(svc, accountRepoMock)
	WithLogger(svc, loggerMock)

	ctx := context.Background()
	ctxTx := context.WithValue(ctx, mySqlExt.CtxSqlTx, &sqlx.Tx{})
	validMerchantID := uuid.NewString()
	validMerchantUUID, _ := uuid.Parse(validMerchantID)

	newAccount := func() *account_model.Account {
		return &account_model.Account{
			UUID:          uuid.New(),
			ReferenceID:   validMerchantUUID,
			Name:          constant.AccountNamePayment,
			HoldedBalance: 100000,
			Currency:      "IDR",
		}
	}

	tests := []struct {
		name        string
		request     *adjustment.HoldMerchantBalanceRequest
		setupMock   func()
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "ERROR: FindMerchantAccountByName database error",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(nil, errors.New("database connection failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: Account not found",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(nil, nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrUnprocessableContent,
		},
		{
			name: "ERROR: BeginTransaction failed",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(newAccount(), nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(nil, errors.New("begin tx failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: UpdateHoldedBalance failed",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(newAccount(), nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(errors.New("update failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: CommitTransaction failed",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(newAccount(), nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(errors.New("commit failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: Zero holded balance results in invalid amount",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				account := newAccount()
				account.HoldedBalance = 0
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(account, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(nil)
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrRequest,
		},
		{
			name: "SUCCESS: Release holded merchant balance",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      50000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(newAccount(), nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Release caps amount to holded balance",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				Amount:      200000,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(newAccount(), nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := svc.ReleaseHoldedMerchantBalance(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.wantErrCode != "" {
					errCode, _ := pkgErrs.ExtractError(err)
					assert.Equal(t, tt.wantErrCode, errCode)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.request.MerchantId, result.MerchantID)
				assert.Equal(t, string(constant.HoldedBalanceTypeRelease), result.Type)
			}

			adjustRepoMock.AssertExpectations(t)
			accountRepoMock.AssertExpectations(t)
			loggerMock.AssertExpectations(t)
		})
	}
}
