package withdrawalService_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateBankTransferStatus(t *testing.T) {
	queue := rmqMock.NewRabbitMQExt(t)
	log := loggerMock.NewILogger(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	ledgerSvc := serviceMocks.NewIOrchestratorService(t)
	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)

	service := New(log, withdrawalRepo, ledgerSvc, nil, nil, WithRabbitMQClient(queue), WithMerchantRepository(merchantRepo))

	request := &routingProcessorModel.BankTransferResponseData{
		UUID: "cac05620-0c06-4f9c-b670-cb769745e426",
		Transaction: &orchestrator_model.AccountTransactionWithUseCase{
			UUID:   uuid.MustParse("f62c1202-bdf8-4e3f-979b-c7093dd9b282"),
			Status: constant.StatusPending,
		},
		Status: constant.SnapCoreBankTransferStatusSuccess,
	}
	withdrawalDetail := &withdrawal.WithdrawalDetailResponse{
		Id: "d4983f09-eb6f-4737-8388-c95f11d4f18e",
		RawMetadata: types.NullJSONText{
			Valid: true, JSONText: []byte(`{"source": "OPEN_API"}`),
		},
	}
	var ptrString *string

	tests := []struct {
		name      string
		request   *routingProcessorModel.BankTransferResponseData
		setupMock func()
		wantError error
	}{
		{
			name:      "SKIP:Transaction is nil",
			request:   &routingProcessorModel.BankTransferResponseData{},
			setupMock: func() { /* Empty */ },
			wantError: nil,
		},
		{
			name: "SKIP:Transaction status is final",
			request: &routingProcessorModel.BankTransferResponseData{
				Transaction: &orchestrator_model.AccountTransactionWithUseCase{
					Status: constant.StatusFailed,
				},
			},
			setupMock: func() {
				log.On(
					"Info", mock.Anything, "Transaction status is final, status cannot be updated. Bank transfer status update ignored", logger.String("status", constant.StatusFailed),
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Get withdrawal detail",
			setupMock: func() {
				withdrawalRepo.On("GetById", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: fmt.Errorf("get withdrawal details: %v", assert.AnError),
		},
		{
			name: "ERROR:Withdrawal detail not found",
			setupMock: func() {
				withdrawalRepo.On("GetById", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: errors.New("withdrawal detail not found"),
		},
		{
			name: "ERROR:Update withdrawal metadata",
			setupMock: func() {
				withdrawalRepo.On("GetById", mock.Anything, mock.Anything).Return(withdrawalDetail, nil)
				withdrawalRepo.On("UpdateMetadataById", mock.Anything, withdrawalDetail.Id, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("update withdrawal metadata: %v", assert.AnError),
		},
		{
			name: "ERROR:Update processor and recon reference",
			setupMock: func() {
				withdrawalRepo.On("UpdateMetadataById", mock.Anything, withdrawalDetail.Id, mock.Anything).Return(nil)
				ledgerSvc.On(
					"UpdateProcessorAndReconReferenceByID", mock.Anything, request.Transaction.UUID.String(), mock.Anything, request.UUID, mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("update processor and recon reference id: %v", assert.AnError),
		},
		{
			name: "ERROR:Update transaction status",
			request: &routingProcessorModel.BankTransferResponseData{
				Transaction:     request.Transaction,
				Status:          constant.SnapCoreBankTransferStatusFailed,
				ResponseCode:    "4030014",
				ResponseMessage: "Insufficient",
			},
			setupMock: func() {
				ledgerSvc.On(
					"UpdateProcessorAndReconReferenceByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil)
				ledgerSvc.On(
					"UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending,
					mock.MatchedBy(func(p *string) bool {
						return p != nil && *p == constant.ReasonTypeInsufficientEscrowFund
					}),
					mock.MatchedBy(func(p *string) bool {
						return p != nil && *p == "Insufficient"
					}),
				).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("update transaction status: %v", assert.AnError),
		},
		{
			name: "SUCCESS:Pending",
			request: &routingProcessorModel.BankTransferResponseData{
				Transaction:     request.Transaction,
				Status:          constant.SnapCoreBankTransferStatusFailed,
				ResponseCode:    "4030014",
				ResponseMessage: "Insufficient",
			},
			setupMock: func() {
				ledgerSvc.On(
					"UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusPending,
					mock.MatchedBy(func(p *string) bool {
						return p != nil && *p == constant.ReasonTypeInsufficientEscrowFund
					}),
					mock.MatchedBy(func(p *string) bool {
						return p != nil && *p == "Insufficient"
					}),
				).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Callback failed",
			setupMock: func() {
				ledgerSvc.On(
					"UpdateStatusAccountTransaction", mock.Anything, mock.Anything, constant.StatusSuccess, ptrString, ptrString,
				).Return(nil)
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Warn", mock.Anything, "Failed to send withdrawal final status", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Completed",
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchant.Merchant{}, nil)
				queue.On("PublishMerchantCallback", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.request == nil {
				test.request = request
			}
			assert.Equal(t, test.wantError, service.UpdateBankTransferStatus(t.Context(), test.request))

			log.AssertExpectations(t)
			queue.AssertExpectations(t)
			ledgerSvc.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			withdrawalRepo.AssertExpectations(t)
		})
	}
}
