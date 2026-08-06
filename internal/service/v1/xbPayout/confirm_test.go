package xbPayoutService

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	rabbitmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	_ "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory" // imported for type checking in mock
)

func TestConfirm(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	mockRabbitmq := rabbitmqMock.NewRabbitMQExt(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)
	feeSvc := serviceMock.NewIFeeService(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)
	statusHistoriesRepo := repositoryMock.NewIStatusHistoriesRepository(t)
	statusHistoriesRepo.On("Insert", c.ValueCtxMockType(), mock.Anything).Return(nil)

	processorReferenceId := "proc001"
	remark := "remark"
	validDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			ProcessorReferenceID: &processorReferenceId,
			Remark:               &remark,
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					ExpiredAt:   time.Now().Add(2 * time.Hour),
					TotalAmount: decimal.NewFromFloat(1_000_000),
				},
			},
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Find payout error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Find payout not found",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Merchant ID is not valid",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MetadataObj: disbursementModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								ExpiredAt: time.Now().Add(time.Hour),
							},
						},
						MerchantID: uuid.NewString(),
					},
				}, nil)
			},
		},
		{
			name:    "ERROR: Payout was expired",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                 uuid.NewString(),
						BeneficiaryAccountNo: c.XBSimulationInsufficientBalanceNumber,
						MetadataObj: disbursementModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								ExpiredAt: time.Now().Add(-time.Hour),
							},
						},
					},
				}, nil)
			},
		},
		{
			name:    "ERROR: Payout has already approved",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(validDisbursement, nil)

				disbursementRepo.On("ApproveInBulk",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
				).Once().Return(c.ErrNoRowsAffected)
			},
		},
		{
			name:    "ERROR: Payout approve service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("ApproveInBulk",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: GetAvailableMerchantBalance service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("ApproveInBulk",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(0.0, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Insufficient amount to confirm and UpdateReasonByIDs service error",
			wantErr: true,
			setupMock: func() {
				orchestratorSvc.On("GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(0.0, nil)

				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Insufficient amount to confirm",
			wantErr: true,
			setupMock: func() {
				orchestratorSvc.On("GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(0.0, nil)

				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: XbCore confirm service error",
			wantErr: true,
			setupMock: func() {
				orchestratorSvc.On("GetAvailableMerchantBalance",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(2_000_000.0, nil)

				xbCoreProcessorRepo.On("ConfirmPayout",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.ConfirmPayoutRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)

				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: UpdateProcessorReferenceIdAndBankReferenceNo error",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("ConfirmPayout",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.ConfirmPayoutRequest"),
				).Return(&xbCoreProcessorModel.ConfirmPayoutResponseData{}, nil)

				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: UpdateReasonByIDs error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateProcessorReferenceIdAndBankReferenceNo",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: FindByReference service error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: PostAccountTransaction error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: PostAccountTransaction fee error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Ledger exists but UpdateStatusAccountTransactionByReferenceID error",
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: Ledger exists and updates successfully",
			wantErr: false,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					c.ValueCtxMockType(),
					c.ArrayStringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Create new ledger transactions",
			wantErr: false,
			setupMock: func() {
				orchestratorSvc.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					c.ValueCtxMockType(),
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				statusHistoriesRepo.On("Insert",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*statusHistoryModel.StatusHistory"),
				).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo,
				WithFeeService(feeSvc),
				WithOrchestratorService(orchestratorSvc),
				WithConfig(cfg),
				WithStatusHistories(statusHistoriesRepo),
				WithRabbitMQClient(mockRabbitmq),
			)
			_, err := svc.Confirm(context.Background(), &xbModel.ConfirmPayoutRequest{
				PayoutId: "payout-123",
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleConfirmationError(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)
	feeSvc := serviceMock.NewIFeeService(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)
	statusHistoriesRepo := repositoryMock.NewIStatusHistoriesRepository(t)
	mockRabbitmq := rabbitmqMock.NewRabbitMQExt(t)

	testDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:        uuid.NewString(),
			MerchantID:  uuid.NewString(),
			CreatedFrom: util.ValueToPtr(constant.DisbursementCreatedFromOpenApi),
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					Uuid:        uuid.NewString(),
					ExpiredAt:   time.Now().Add(2 * time.Hour),
					TotalAmount: decimal.NewFromFloat(1_000_000),
				},
				FeeDetail: feeModel.FeeMetadataObject{
					FinalAmount: 5000.0,
				},
			},
		},
	}

	testCases := []struct {
		name      string
		request   *xbModel.ConfirmPayoutRequest
		wantErr   bool
		setupMock func()
	}{
		{
			name: "ERROR: UpdateReasonByIDs fails",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: RecordStatusHistory fails but does not return error",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: false,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR: FindByReference service error",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: Ledger exists but UpdateStatusAccountTransactionByReferenceID fails",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: true,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID",
					mock.Anything,
					testDisbursement.UUID,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS: Actor defaults to system user when ApprovedBy is empty",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "",
			},
			wantErr: false,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(nil, nil)
			},
		},
		{
			name: "SUCCESS: Ledger does not exist",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: false,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(nil, nil)
			},
		},
		{
			name: "SUCCESS: Ledger exists and updates successfully",
			request: &xbModel.ConfirmPayoutRequest{
				PayoutId:   "payout-123",
				MerchantId: testDisbursement.MerchantID,
				ApprovedBy: "user-123",
			},
			wantErr: false,
			setupMock: func() {
				disbursementRepo.On("UpdateReasonByIDs",
					mock.Anything,
					[]string{"payout-123"},
					c.XbDisbursementReasonTypeError,
					c.XbDisbursementReasonDescError,
				).Once().Return(nil)

				statusHistoriesRepo.On("Insert",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				orchestratorSvc.On("FindByReference",
					mock.Anything,
					testDisbursement.UUID,
					orchestratorModel.TypeDisbursement,
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				orchestratorSvc.On("UpdateStatusAccountTransactionByReferenceID",
					mock.Anything,
					testDisbursement.UUID,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, disbursementRepo, nil, xbCoreProcessorRepo,
				WithFeeService(feeSvc),
				WithOrchestratorService(orchestratorSvc),
				WithConfig(cfg),
				WithStatusHistories(statusHistoriesRepo),
				WithRabbitMQClient(mockRabbitmq),
			)

			// Access the private method directly since we're in the same package
			xbSvc := svc.(*xbPayoutService)
			err := xbSvc.handleConfirmationError(context.Background(), tc.request, testDisbursement)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
