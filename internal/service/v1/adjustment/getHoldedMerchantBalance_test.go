package adjustment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
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
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func TestGetHoldedMerchantBalance(t *testing.T) {
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	accountRepoMock := repoMocks.NewIAccountRepository(t)
	loggerMock := loggerMocks.NewILogger(t)

	svc := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithAccountRepository(svc, accountRepoMock)
	WithLogger(svc, loggerMock)

	ctx := context.Background()
	validMerchantID := uuid.NewString()
	validMerchantUUID, _ := uuid.Parse(validMerchantID)

	validAccount := &account_model.Account{
		UUID:          uuid.New(),
		ReferenceID:   validMerchantUUID,
		Name:          constant.AccountNamePayment,
		HoldedBalance: 75000,
		Currency:      "IDR",
	}

	tests := []struct {
		name        string
		request     *adjustment.GetHoldedMerchantBalanceRequest
		setupMock   func()
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "ERROR: FindMerchantAccountByName database error",
			request: &adjustment.GetHoldedMerchantBalanceRequest{
				MerchantId:  validMerchantID,
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
			request: &adjustment.GetHoldedMerchantBalanceRequest{
				MerchantId:  validMerchantID,
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
			name: "SUCCESS: Get holded merchant balance",
			request: &adjustment.GetHoldedMerchantBalanceRequest{
				MerchantId:  validMerchantID,
				AccountType: constant.AccountNamePayment,
			},
			setupMock: func() {
				accountRepoMock.On(
					"FindMerchantAccountByName", mock.Anything, validMerchantUUID, constant.AccountNamePayment,
				).Once().Return(validAccount, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := svc.GetHoldedMerchantBalance(ctx, tt.request)

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
				assert.Equal(t, tt.request.AccountType, result.AccountType)
				assert.Equal(t, validAccount.HoldedBalance, result.Amount)
			}

			accountRepoMock.AssertExpectations(t)
		})
	}
}
