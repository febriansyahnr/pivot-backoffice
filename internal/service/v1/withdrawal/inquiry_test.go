package withdrawalService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInquiryTransaction(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	service := New(
		logger, withdrawalRepo, orchestratorSvc, nil, nil, WithSnapCoreRepository(snapCoreRepo),
	)

	withdrawalId := uuid.NewString()
	withdrawalDetails := &withdrawal.WithdrawalDetailResponse{
		Id:                     uuid.NewString(),
		CreatedAt:              time.Now().UTC(),
		CreatedBy:              "John",
		Amount:                 20_000,
		Status:                 c.StatusPending,
		BankReferenceNo:        "1234",
		BeneficiaryBankName:    "BANK TEST",
		BeneficiaryAccountNo:   "000001",
		BeneficiaryAccountName: "Hendru",
		TransactionID:          uuid.NewString(),
		ExternalID:             uuid.NewString(),
	}
	result := &withdrawal.InquiryTransactionResponse{
		WithdrawalDetailResponse: withdrawalDetails,
		Status:                   c.StatusPending,
		ReasonType:               c.ReasonTypeInsufficientEscrowFund,
		ReasonDescription:        "TEST",
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.InquiryTransactionResponse
	}{
		{
			name: "ERROR:Withdrawal data is not found",
			setupMock: func() {
				withdrawalRepo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrDataNotFound), // NOSONAR
		},
		{
			name: "ERROR:Withdrawal status is final",
			setupMock: func() {
				withdrawalRepo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(&withdrawal.WithdrawalDetailResponse{
					Status: c.StatusSuccess,
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrTransactionAlreadyInFinalStatus), // NOSONAR
		},
		{
			name: "ERROR:Empty external id",
			setupMock: func() {
				withdrawalRepo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(&withdrawal.WithdrawalDetailResponse{
					Status: c.StatusPending,
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("external id for bank transfer not found")), // NOSONAR
		},
		{
			name: "ERROR:Find bank transfer by external id",
			setupMock: func() {
				withdrawalRepo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Return(withdrawalDetails, nil)

				snapCoreRepo.On(
					"FindBankTransferByExternalID", c.ValueCtxMockType(), withdrawalDetails.ExternalID, false,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Update metadata by id",
			setupMock: func() {
				snapCoreRepo.On(
					"FindBankTransferByExternalID", c.ValueCtxMockType(), withdrawalDetails.ExternalID, false,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					Status:          c.SnapCoreBankTransferStatusSuccess,
					BankReferenceNo: "123456",
				}, nil)
				withdrawalRepo.On(
					"UpdateMetadataById", c.ValueCtxMockType(), withdrawalId, mock.AnythingOfType("*withdrawal.Metadata"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Update status account transaction",
			setupMock: func() {
				snapCoreRepo.On(
					"FindBankTransferByExternalID", c.ValueCtxMockType(), withdrawalDetails.ExternalID, false,
				).Return(&snapCoreModel.BankTransferResponseData{
					Status:          c.SnapCoreBankTransferStatusFailed,
					ResponseCode:    "4030014",
					ResponseMessage: "TEST",
				}, nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), withdrawalDetails.TransactionID, result.Status, &result.ReasonType, &result.ReasonDescription,
				).Once().Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), withdrawalDetails.TransactionID, result.Status, &result.ReasonType, &result.ReasonDescription,
				).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			request := &withdrawal.InquiryTransactionRequest{Id: withdrawalId}

			result, err := service.InquiryTransaction(context.Background(), request)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.Equal(t, test.wantErr, err)

			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				test.wantResult.UpdatedAt = result.UpdatedAt
				assert.Equal(t, test.wantResult, result)
			}
		})
	}
}
