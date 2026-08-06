package transferService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"errors"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReverseTransfer(t *testing.T) {
	merchantID := uuid.New()
	parentMerchantID := uuid.New()
	sourceAccountID := uuid.New()
	recipientAccountID := uuid.New()

	existingTransfer := &transfer.Transfer{
		UUID:         uuid.New(),
		MerchantID:   parentMerchantID,
		RecipientID:  merchantID,
		ReferenceID:  "ref-123",
		Amount:       10_000,
		TransferType: constant.MoneyFlowDirect,
		Currency:     constant.CurrencyIDR,
	}

	validRequest := &transfer.ReverseTransferRequest{
		ReferenceID:      "ref-123",
		MerchantID:       merchantID.String(),
		ParentMerchantID: parentMerchantID.String(),
		Remarks:          "reverse test",
	}

	testCases := []struct {
		name      string
		request   *transfer.ReverseTransferRequest
		mockSetup func(
			accountSvc *mockSvc.IAccountService,
			ledgerSvc *mockSvc.ILedgerService,
			platformFeeSvc *mockSvc.IPlatformFeeService,
			transferRepo *mockRepo.ITransferRepository,
		)
		expectErr bool
	}{
		{
			name:    "SUCCESS: reverse transfer",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
					merchantID:       {UUID: recipientAccountID},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				platformFeeSvc.On("ReverseMerchantFee", mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name:    "ERROR: GetByReferenceID fails",
			request: validRequest,
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(nil, errors.New("db error"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: transfer not found",
			request: validRequest,
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(nil, nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: create reverse transfer fails",
			request: validRequest,
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(errors.New("insert error"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: get merchant accounts fails",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("account error"))
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: sender account not found",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					merchantID: {UUID: recipientAccountID},
				}, nil)
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: recipient account not found",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
				}, nil)
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: record ledger transaction fails",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
					merchantID:       {UUID: recipientAccountID},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ledger error"))
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: update status to failed after ledger error",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
					merchantID:       {UUID: recipientAccountID},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ledger error"))
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(errors.New("update error"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: update status to success fails",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
					merchantID:       {UUID: recipientAccountID},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(errors.New("update error"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: reverse merchant fee fails",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, transferRepo *mockRepo.ITransferRepository) {
				transferRepo.On("GetByReferenceID", mock.Anything, validRequest.MerchantID, validRequest.ParentMerchantID, validRequest.ReferenceID).Return(existingTransfer, nil)
				transferRepo.On("Create", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{
					parentMerchantID: {UUID: sourceAccountID},
					merchantID:       {UUID: recipientAccountID},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				transferRepo.On("Update", mock.Anything, mock.AnythingOfType("*transfer.Transfer")).Return(nil)
				platformFeeSvc.On("ReverseMerchantFee", mock.Anything, mock.Anything).Return(errors.New("fee reversal error"))
			},
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accSvc := mockSvc.NewIAccountService(t)
			ledgerSvc := mockSvc.NewILedgerService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			transferRepo := mockRepo.NewITransferRepository(t)
			platformFeeSvc := mockSvc.NewIPlatformFeeService(t)
			merchantSvc := mockSvc.NewIMerchantService(t)

			tc.mockSetup(accSvc, ledgerSvc, platformFeeSvc, transferRepo)
			svc := New(mockLogger, ledgerSvc, accSvc, platformFeeSvc, merchantSvc, transferRepo)
			resp, err := svc.ReverseTransfer(context.Background(), tc.request)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, resp)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
