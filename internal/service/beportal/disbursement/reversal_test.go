package disbursementService_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReversal(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	disbursementWithoutFee := disbursementModel.DisbursementForReversal{
		Id:         uuid.NewString(),
		MerchantId: uuid.NewString(),
		Status:     c.DisbursementStatusApproved,
		Amount:     25_000.00,
		Currency:   "IDR",
		Transaction: disbursementModel.TransactionMetadataForReversal{
			Id:     uuid.NewString(),
			Status: c.StatusSuccess,
			Amount: 25_000.00,
		},
	}
	disbursementWithIndirectFee := disbursementModel.DisbursementForReversal{
		Id:         uuid.NewString(),
		MerchantId: uuid.NewString(),
		Status:     c.DisbursementStatusApproved,
		Amount:     25_000.00,
		Currency:   "IDR",
		FeeAmount:  2_500.00,
		Transaction: disbursementModel.TransactionMetadataForReversal{
			Id:     uuid.NewString(),
			Status: c.StatusSuccess,
			Amount: 25_000.00,
		},
		Fee: disbursementModel.TransactionMetadataForReversal{
			Id:     uuid.NewString(),
			Status: c.StatusPending,
			Amount: 2_500.00,
			Metadata: feeModel.FeeMetadataObject{
				DeductionType: c.MerchantFeeDeductionTypeAutomated,
			},
		},
	}
	disbursementWithDirectFeeSuccess := disbursementModel.DisbursementForReversal{
		Id:         uuid.NewString(),
		MerchantId: uuid.NewString(),
		Status:     c.DisbursementStatusApproved,
		Amount:     25_000.00,
		Currency:   "IDR",
		FeeAmount:  2_500.00,
		Transaction: disbursementModel.TransactionMetadataForReversal{
			Id:     uuid.NewString(),
			Status: c.StatusSuccess,
			Amount: 25_000.00,
		},
		Fee: disbursementModel.TransactionMetadataForReversal{
			Id:     uuid.NewString(),
			Status: c.StatusSuccess,
			Amount: 2_500.00,
			Metadata: feeModel.FeeMetadataObject{
				DeductionType: c.MerchantFeeDeductionTypeDirect,
			},
		},
	}

	merchantWithoutParent := &merchantModel.Merchant{
		UUID: uuid.NewString(),
	}
	parentMerchantId := uuid.NewString()
	merchantWithParent := &merchantModel.Merchant{
		UUID:     uuid.NewString(),
		ParentID: sql.NullString{String: parentMerchantId, Valid: true},
	}

	newSvc := func() (service.IDisbursementService, *serviceMocks.IOrchestratorService, *repoMocks.IDisbursementRepository, *repoMocks.IAccountTransactionRepository, *serviceMocks.IMerchantService, *serviceMocks.ITransferService) {
		orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
		disbursementRepo := repoMocks.NewIDisbursementRepository(t)
		accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)
		merchantSvc := serviceMocks.NewIMerchantService(t)
		transferSvc := serviceMocks.NewITransferService(t)

		svc := New(
			&config.Config{}, logger, nil, disbursementRepo, nil, nil,
			WithOrchestratorService(orchestratorSvc),
			WithAccountTransactionRepository(accountTransactionRepo),
			WithMerchantService(merchantSvc),
			WithTransferService(transferSvc),
		)
		return svc, orchestratorSvc, disbursementRepo, accountTransactionRepo, merchantSvc, transferSvc
	}

	tests := []struct {
		name       string
		setupMock  func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, accountTransactionRepo *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, transferSvc *serviceMocks.ITransferService)
		wantErr    string
		wantResult *disbursementModel.ReversalTransactionResp
	}{
		{
			name: "ERROR:Find disbursement data",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, _ *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Disbursement data not found",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, _ *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "disbursement data not found",
		},
		{
			name: "ERROR:Disbursement status WAITING",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, _ *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementForReversal{
					Status: c.DisbursementStatusWaiting,
				}, nil)
			},
			wantErr: "disbursement status must be approved",
		},
		{
			name: "ERROR:Disbursement has been REVERSAL",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, _ *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementForReversal{
					Status: c.DisbursementStatusApproved, ReasonType: sql.NullString{String: c.ReasonTypeReversal},
				}, nil)
			},
			wantErr: "disbursement has been REVERSAL",
		},
		{
			name: "ERROR:Pending transaction",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, _ *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementForReversal{
					Status: c.DisbursementStatusApproved, Transaction: disbursementModel.TransactionMetadataForReversal{Status: c.StatusPending},
				}, nil)
			},
			wantErr: "transaction must be have SUCCESS status",
		},
		{
			name: "ERROR:Getting merchant data",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Merchant data not found",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)
			},
			wantErr: "disbursement merchant data not found",
		},
		{
			name: "ERROR:Begin transaction",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Cancel indirect transaction fee",
			setupMock: func(_ *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, accountTransactionRepo *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithIndirectFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				accountTransactionRepo.On(
					"CancelIndirectTransactionFee", c.ValueCtxMockType(), c.StringMockType(), c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Post account transaction",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				orchestratorSvc.On(
					"PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Update disbursement reason",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				disbursementRepo.On(
					"UpdateReversalTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Commit transaction",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				disbursementRepo.On("UpdateReversalTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS: without parent merchant",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On(
					"FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&disbursementWithoutFee, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithoutParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				disbursementRepo.On("UpdateReversalTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantResult: &disbursementModel.ReversalTransactionResp{
				DisbursementId: disbursementWithoutFee.Id,
				ReversalAmount: disbursementWithoutFee.Amount,
			},
		},
		{
			name: "SUCCESS: with parent merchant and fee reversal",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, transferSvc *serviceMocks.ITransferService) {
				disbursementRepo.On("FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).Once().Return(&disbursementWithDirectFeeSuccess, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				// Parent fee account transaction
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				// Reverse transfer
				transferSvc.On("ReverseTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*transfer.ReverseTransferRequest")).Once().Return(nil, nil)
				// Main account transaction
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				disbursementRepo.On("UpdateReversalTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).Once().Return(nil)
				disbursementRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantResult: &disbursementModel.ReversalTransactionResp{
				DisbursementId: disbursementWithDirectFeeSuccess.Id,
				ReversalAmount: disbursementWithDirectFeeSuccess.Amount,
			},
		},
		{
			name: "ERROR: reverse transfer fails for sub-merchant",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, transferSvc *serviceMocks.ITransferService) {
				disbursementRepo.On("FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).Once().Return(&disbursementWithDirectFeeSuccess, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(nil)
				transferSvc.On("ReverseTransfer", c.ValueCtxMockType(), mock.AnythingOfType("*transfer.ReverseTransferRequest")).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR: parent fee post account transaction fails",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService, disbursementRepo *repoMocks.IDisbursementRepository, _ *repoMocks.IAccountTransactionRepository, merchantSvc *serviceMocks.IMerchantService, _ *serviceMocks.ITransferService) {
				disbursementRepo.On("FindForReversalDisbursementById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).Once().Return(&disbursementWithDirectFeeSuccess, nil)
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), mock.Anything).Once().Return(merchantWithParent, nil)
				disbursementRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(ctx, mySqlExt.CtxSqlTx, nil), nil)
				// Parent fee PostAccountTransaction — fail
				orchestratorSvc.On("PostAccountTransaction", c.ValueCtxMockType(), c.PtrCreateAccTransactionReqMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				disbursementRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, orchestratorSvc, disbursementRepo, accountTransactionRepo, merchantSvc, transferSvc := newSvc()
			test.setupMock(orchestratorSvc, disbursementRepo, accountTransactionRepo, merchantSvc, transferSvc)

			result, err := svc.Reversal(ctx, &disbursementModel.ReversalTransactionReq{})
			if test.wantErr == "" {
				require.NoError(t, err)
				test.wantResult.Id = result.Id
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, test.wantErr)
			}
			assert.Equal(t, test.wantResult, result)
		})
	}
}
