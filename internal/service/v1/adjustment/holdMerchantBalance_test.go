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
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func TestHoldMerchantBalance(t *testing.T) {
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	accountRepoMock := repoMocks.NewIAccountRepository(t)
	orchestratorMock := serviceMocks.NewIOrchestratorService(t)
	loggerMock := loggerMocks.NewILogger(t)

	loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	svc := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithOrchestratorService(svc, orchestratorMock)
	WithLogger(svc, loggerMock)
	WithAccountRepository(svc, accountRepoMock)

	ctx := context.Background()
	ctxTx := context.WithValue(ctx, mySqlExt.CtxSqlTx, &sqlx.Tx{})

	validMerchant := &merchantModel.Merchant{
		UUID: uuid.NewString(),
		Name: "Test Merchant",
	}

	validRequest := &adjustment.HoldMerchantBalanceRequest{
		MerchantId:  uuid.NewString(),
		Amount:      50000,
		AccountType: constant.AccountNamePayment,
	}

	validAccount := &accountModel.Account{
		UUID:          uuid.New(),
		ReferenceID:   uuid.MustParse(validRequest.MerchantId),
		Name:          constant.AccountNamePayment,
		HoldedBalance: 0,
		Currency:      "IDR",
	}

	tests := []struct {
		name         string
		request      *adjustment.HoldMerchantBalanceRequest
		setupMock    func()
		wantErr      bool
		wantErrCode  string
		validateResp func(*testing.T, *adjustment.HoldMerchantBalanceResponse)
	}{
		{
			name:    "ERROR: FindMerchantByID database error",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(nil, errors.New("database connection failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Merchant not found",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(nil, nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrUnprocessableContent,
		},
		{
			name:    "ERROR: GetAvailableMerchantBalance database error",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(0.0, errors.New("balance service unavailable"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Insufficient balance",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(10000.0, nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrUnprocessableContent,
		},
		{
			name:    "ERROR: FindMerchantAccountByName database error",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Account not found",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(nil, nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrUnprocessableContent,
		},
		{
			name:    "ERROR: Begin transaction failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(nil, errors.New("begin transaction failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: Invalid amount - zero amount",
			request: &adjustment.HoldMerchantBalanceRequest{
				MerchantId:  uuid.NewString(),
				Amount:      0,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, mock.Anything, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrRequest,
		},
		{
			name:    "ERROR: UpdateHoldedBalance failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				accountRepoMock.On(
					"UpdateHoldedBalance", mock.Anything, mock.AnythingOfType("*account_model.Account"),
				).Once().Return(errors.New("update holded balance failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: CommitTransaction failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
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
			name:    "SUCCESS: Hold merchant balance",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				orchestratorMock.On(
					"GetAvailableMerchantBalance", mock.Anything, validRequest.MerchantId, constant.AccountNamePayment,
				).Once().Return(100000.0, nil)
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, mock.AnythingOfType("uuid.UUID"), constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
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
			validateResp: func(t *testing.T, resp *adjustment.HoldMerchantBalanceResponse) {
				assert.Equal(t, validRequest.MerchantId, resp.MerchantID)
				assert.Equal(t, validRequest.Amount, resp.Amount)
				assert.Equal(t, string(constant.HoldedBalanceTypeHold), resp.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := svc.HoldMerchantBalance(ctx, tt.request)

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

				if tt.validateResp != nil {
					tt.validateResp(t, result)
				}
			}

			merchantRepoMock.AssertExpectations(t)
			adjustRepoMock.AssertExpectations(t)
			orchestratorMock.AssertExpectations(t)
			accountRepoMock.AssertExpectations(t)
		})
	}
}
