package withdrawalService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryTransaction(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	accountTrxRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := New(
		logger, withdrawalRepo, orchestratorSvc, nil, nil,
		WithSnapCoreRepository(snapCoreRepo),
		WithAccountTransactionRepository(accountTrxRepo),
	)

	withdrawalID := uuid.NewString()
	merchantID := uuid.NewString()
	externalID := uuid.New()

	wd := &withdrawal.Withdrawal{
		Id:                     withdrawalID,
		MerchantId:             merchantID,
		BeneficiaryBankCode:    "014",
		BeneficiaryBankName:    "BCA",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryAccountName: "Test User",
		Amount:                 100_000,
		Currency:               "IDR",
		CreatedAt:              time.Now().UTC(),
		RawMetadata: types.NullJSONText{
			Valid: true, JSONText: []byte(`{"source":"DASHBOARD","withdrawType":"BANK_TRANSFER"}`),
		},
		Metadata: withdrawal.Metadata{
			Source:       "DASHBOARD",
			WithdrawType: c.WithdrawalDestBankTransfer,
		},
	}

	accountTrx := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:   externalID,
		Status: c.StatusFailed,
	}

	request := &withdrawal.RetryTransactionRequest{
		WithdrawalID: withdrawalID,
		MerchantID:   merchantID,
	}

	tests := []struct {
		name      string
		request   *withdrawal.RetryTransactionRequest
		setupMock func()
		wantErr   error
	}{
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
			name:    "ERROR: Balance transfer withdrawal not retryable",
			request: request,
			setupMock: func() {
				balanceTransferWd := &withdrawal.Withdrawal{
					Id:         withdrawalID,
					MerchantId: merchantID,
					Metadata: withdrawal.Metadata{
						WithdrawType: c.WithdrawalDestBalanceTransfer,
					},
				}
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(balanceTransferWd, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("retry is not supported for balance transfer withdrawals")),
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
			name:    "ERROR: CheckAllowedToRetry returns error",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				snapCoreRepo.On(
					"CheckAllowedToRetry", c.ValueCtxMockType(), snapCoreModel.CheckAllowedToRetryRequest{
						ExternalID: externalID.String(),
						MerchantId: merchantID,
						Force:      false,
					},
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name:    "ERROR: CheckAllowedToRetry not allowed",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				snapCoreRepo.On(
					"CheckAllowedToRetry", c.ValueCtxMockType(), snapCoreModel.CheckAllowedToRetryRequest{
						ExternalID: externalID.String(),
						MerchantId: merchantID,
						Force:      false,
					},
				).Once().Return(&snapCoreModel.CheckAllowedToRetryResponse{
					Allowed: false,
					Reason:  "bank transfer metadata async status is not done",
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrRequest, errors.New("bank transfer metadata async status is not done")),
		},
		{
			name:    "ERROR: BankTransfer call fails",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				snapCoreRepo.On(
					"CheckAllowedToRetry", c.ValueCtxMockType(), snapCoreModel.CheckAllowedToRetryRequest{
						ExternalID: externalID.String(),
						MerchantId: merchantID,
						Force:      false,
					},
				).Once().Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: true}, nil)
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.BankTransferRequest"), mock.AnythingOfType("*snapCoreModel.BankTransferHeaderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), externalID.String(), c.StatusFailed, mock.AnythingOfType("*string"), mock.AnythingOfType("*string"),
				).Once().Return(nil)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS: BypassProcessorCheck skips CheckAllowedToRetry",
			request: &withdrawal.RetryTransactionRequest{
				WithdrawalID:         withdrawalID,
				MerchantID:           merchantID,
				BypassProcessorCheck: true,
			},
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				// No CheckAllowedToRetry call expected
				snapCoreResp := &snapCoreModel.BankTransferResponseData{
					UUID:       uuid.NewString(),
					ExternalID: externalID.String(),
					Status:     c.SnapCoreBankTransferStatusSuccess,
				}
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.BankTransferRequest"), mock.AnythingOfType("*snapCoreModel.BankTransferHeaderRequest"),
				).Once().Return(snapCoreResp, nil)
				withdrawalRepo.On(
					"UpdateMetadataById", c.ValueCtxMockType(), withdrawalID, mock.AnythingOfType("*withdrawal.Metadata"),
				).Once().Return(nil)
				orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), externalID.String(), c.SnapCoreProcessor, snapCoreResp.UUID, mock.AnythingOfType("string"),
				).Once().Return(nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), externalID.String(), c.StatusSuccess, (*string)(nil), (*string)(nil),
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS: ForceRetry passes Force flag to snap-core",
			request: &withdrawal.RetryTransactionRequest{
				WithdrawalID: withdrawalID,
				MerchantID:   merchantID,
				ForceRetry:   true,
			},
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				snapCoreRepo.On(
					"CheckAllowedToRetry", c.ValueCtxMockType(), snapCoreModel.CheckAllowedToRetryRequest{
						ExternalID: externalID.String(),
						MerchantId: merchantID,
						Force:      true,
					},
				).Once().Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: true}, nil)
				snapCoreResp := &snapCoreModel.BankTransferResponseData{
					UUID:       uuid.NewString(),
					ExternalID: externalID.String(),
					Status:     c.SnapCoreBankTransferStatusPending,
				}
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.BankTransferRequest"), mock.AnythingOfType("*snapCoreModel.BankTransferHeaderRequest"),
				).Once().Return(snapCoreResp, nil)
				withdrawalRepo.On(
					"UpdateMetadataById", c.ValueCtxMockType(), withdrawalID, mock.AnythingOfType("*withdrawal.Metadata"),
				).Once().Return(nil)
				orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), externalID.String(), c.SnapCoreProcessor, snapCoreResp.UUID, mock.AnythingOfType("string"),
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Normal retry with allowed response code",
			request: request,
			setupMock: func() {
				withdrawalRepo.On(
					"FindById", c.ValueCtxMockType(), withdrawalID, merchantID,
				).Once().Return(wd, nil)
				accountTrxRepo.On(
					"FindByReference", c.ValueCtxMockType(), withdrawalID, c.TypeWithdrawal,
				).Once().Return(accountTrx, nil)
				snapCoreRepo.On(
					"CheckAllowedToRetry", c.ValueCtxMockType(), snapCoreModel.CheckAllowedToRetryRequest{
						ExternalID: externalID.String(),
						MerchantId: merchantID,
						Force:      false,
					},
				).Once().Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: true}, nil)
				snapCoreResp := &snapCoreModel.BankTransferResponseData{
					UUID:               uuid.NewString(),
					ExternalID:         externalID.String(),
					Status:             c.SnapCoreBankTransferStatusSuccess,
					BankReferenceNo:    "REF123",
					PartnerReferenceNo: "PARTNER456",
				}
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.BankTransferRequest"), mock.AnythingOfType("*snapCoreModel.BankTransferHeaderRequest"),
				).Once().Return(snapCoreResp, nil)
				withdrawalRepo.On(
					"UpdateMetadataById", c.ValueCtxMockType(), withdrawalID, mock.AnythingOfType("*withdrawal.Metadata"),
				).Once().Return(nil)
				orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), externalID.String(), c.SnapCoreProcessor, snapCoreResp.UUID, mock.AnythingOfType("string"),
				).Once().Return(nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), externalID.String(), c.StatusSuccess, (*string)(nil), (*string)(nil),
				).Once().Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.RetryTransaction(context.Background(), test.request)
			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
