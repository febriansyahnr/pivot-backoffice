package refundMoneyFlowService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefundMoneyFlowService_CreateTransactions(t *testing.T) {
	mockTime := time.Now()
	ctx := context.Background()

	tests := []struct {
		name          string
		request       *ledgerModel.CreateNewLedgerEntryRequest
		mockSetup     func(*mocks.IAccountTransactionRepository)
		expectedError error
	}{
		{
			name: "success with refund transactions",
			request: &ledgerModel.CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-1",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				TransactionTimestamp: mockTime,
				Fee: ledgerModel.FeeRequest{
					Amount: 0.0,
				},
				RefundConfig: ledgerModel.RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			mockSetup: func(repo *mocks.IAccountTransactionRepository) {
				repo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "error when bulk insert fails",
			request: &ledgerModel.CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-2",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				TransactionTimestamp: mockTime,
				Fee: ledgerModel.FeeRequest{
					Amount: 0.0,
				},
				RefundConfig: ledgerModel.RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			mockSetup: func(repo *mocks.IAccountTransactionRepository) {
				repo.On("BulkInsert", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: pkgErrors.New(response.HttpErrInternal, constant.ErrStoreLedgerEntry),
		},
		{
			name: "error when creating refund transactions - invalid usecase",
			request: &ledgerModel.CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-3",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              "INVALID_USECASE",
				TransactionTimestamp: mockTime,
				Fee: ledgerModel.FeeRequest{
					Amount: 0.0,
				},
				RefundConfig: ledgerModel.RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			mockSetup: func(repo *mocks.IAccountTransactionRepository) {
				// No mock setup needed as error occurs before repo call
			},
			expectedError: pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &mocks.IAccountTransactionRepository{}
			mockAccountSvc := &serviceMocks.IAccountService{}
			mockLedgerSvc := &serviceMocks.ILedgerService{}
			mockMerchantSvc := &serviceMocks.IMerchantService{}
			mockLogger := logger.NewSlogger(logger.Config{})

			tt.mockSetup(mockRepo)

			// Create service
			service := New(
				mockLogger,
				mockRepo,
				mockAccountSvc,
				mockLedgerSvc,
				mockMerchantSvc,
			)

			// Execute
			err := service.CreateTransactions(ctx, tt.request)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
