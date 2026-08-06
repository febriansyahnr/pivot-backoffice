package withdrawalService_test

import (
	"context"
	"errors"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestChangeStatusWithdrawal(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	accountTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

	svc := New(
		logger, withdrawalRepo, orchestratorSvc, nil, nil,
		WithAccountTransactionRepository(accountTrxRepo),
	)

	withdrawalID := uuid.NewString()
	merchantID := uuid.NewString()
	externalID := uuid.New()

	wd := &withdrawal.Withdrawal{
		Id:         withdrawalID,
		MerchantId: merchantID,
		Metadata: withdrawal.Metadata{
			WithdrawType: c.WithdrawalDestBankTransfer,
		},
	}

	accountTrx := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:   externalID,
		Status: c.StatusFailed,
	}

	reasonType := "OTHER"
	reasonDescription := "User requested cancellation"
	request := &withdrawal.WithdrawalChangeStatusRequest{
		WithdrawalID:      withdrawalID,
		MerchantID:        merchantID,
		Status:            c.StatusFailed,
		ReasonType:        &reasonType,
		ReasonDescription: &reasonDescription,
	}

	tests := []struct {
		name       string
		request    *withdrawal.WithdrawalChangeStatusRequest
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.WithdrawalChangeStatusResponse
	}{
		{
			name:    "ERROR: FindById database error",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name:    "ERROR: Withdrawal not found",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrNotFound, c.ErrDataNotFound),
		},
		{
			name:    "ERROR: Balance transfer withdrawal",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(&withdrawal.Withdrawal{
					Id:         withdrawalID,
					MerchantId: merchantID,
					Metadata: withdrawal.Metadata{
						WithdrawType: c.WithdrawalDestBalanceTransfer,
					},
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("retry is not supported for balance transfer withdrawals")),
		},
		{
			name:    "ERROR: FindByReference database error",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name:    "ERROR: Account transaction not found",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrNotFound, errors.New("account transaction not found for this withdrawal")),
		},
		{
			name:    "ERROR: Account transaction already SUCCESS",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   externalID,
					Status: c.StatusSuccess,
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrTransactionAlreadyInFinalStatus),
		},
		{
			name:    "ERROR: UpdateStatusAccountTransaction returns error",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), externalID.String(), request.Status, request.ReasonType, request.ReasonDescription,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name:    "SUCCESS: Cancel withdrawal",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), externalID.String(), request.Status, request.ReasonType, request.ReasonDescription,
				).Once().Return(nil)
			},
			wantResult: &withdrawal.WithdrawalChangeStatusResponse{
				ID:         withdrawalID,
				MerchantID: merchantID,
				Status:     c.StatusFailed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := svc.ChangeStatusWithdrawal(context.Background(), test.request)
			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			}
		})
	}
}
