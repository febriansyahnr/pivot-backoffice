package transferService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransfer(t *testing.T) {
	sourceId := uuid.New()
	recipientId := uuid.New()
	parentId := uuid.New()
	merchantAccounts := map[uuid.UUID]*account_model.Account{
		sourceId: {
			UUID: uuid.New(),
		},
		recipientId: {
			UUID: uuid.New(),
		},
		parentId: {
			UUID: uuid.New(),
		},
	}
	validRequest := &transfer.TransferRequest{
		SourceMerchantID: parentId,
		RecipientID:      recipientId.String(),
		ReferenceID:      "reference-id",
		TransferType:     constant.MoneyFlowDirect,
		Amount:           10,
		Remarks:          "test",
		ParentMerchantID: parentId,
	}

	testCases := []struct {
		name      string
		mockSetup func(
			accountSvc *mockSvc.IAccountService,
			ledgerSvc *mockSvc.ILedgerService,
			platformFeeSvc *mockSvc.IPlatformFeeService,
			merchantSvc *mockSvc.IMerchantService,
			mockRepo *mockRepo.ITransferRepository,
		)
		expectErr bool
		request   *transfer.TransferRequest
	}{
		{
			name:    "SUCCESS: Transfer request from merchant account",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts", mock.Anything, mock.Anything, constant.StringMockType(),
				).Return(
					map[uuid.UUID]*account_model.Account{recipientId: {UUID: uuid.New()}, parentId: {UUID: uuid.New()}}, nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, constant.StringMockType(), mock.Anything).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransferFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransactionFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Transfer request from merchant account with fee",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts", mock.Anything, mock.Anything, constant.StringMockType(),
				).Return(
					map[uuid.UUID]*account_model.Account{recipientId: {UUID: uuid.New()}, parentId: {UUID: uuid.New()}}, nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

				ledgerSvc.On("RecordTransaction", mock.Anything, constant.StringMockType(), mock.Anything).Once().Return(nil)
				platformFeeSvc.On("ApplyMerchantTransferFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransactionFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)

			},
		},
		{
			name:    "SUCCESS: Transfer request from sub merchant account",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts", mock.Anything, mock.Anything, constant.StringMockType(),
				).Return(merchantAccounts, nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

				ledgerSvc.On(
					"RecordTransaction", mock.Anything, constant.StringMockType(), mock.Anything,
				).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransferFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransactionFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Transfer request from sub merchant account with fee",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts", mock.Anything, mock.Anything, constant.StringMockType(),
				).Return(merchantAccounts, nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)

				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

				ledgerSvc.On(
					"RecordTransaction",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
				).Once().Return(
					nil,
				)

				platformFeeSvc.On("ApplyMerchantTransferFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				platformFeeSvc.On("ApplyMerchantTransactionFee", constant.ValueCtxMockType(), mock.Anything).Return(nil)

			},
		},
		{
			name:    "ERROR: Find merchant by recipient id",
			request: validRequest,
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, _ *mockRepo.ITransferRepository) {
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(nil, assert.AnError)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Recipient id not found",
			request: validRequest,
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, _ *mockRepo.ITransferRepository) {
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(nil, nil)
			},
			expectErr: true,
		},

		{
			name: "ERROR: Invalid Request",
			request: &transfer.TransferRequest{
				RecipientID: recipientId.String(),
			},
			mockSetup: func(_ *mockSvc.IAccountService, _ *mockSvc.ILedgerService, _ *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, _ *mockRepo.ITransferRepository) {
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Get merchant accounts",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts", mock.Anything, mock.Anything, constant.StringMockType(),
				).Return(nil, errors.New("error"))
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Error Get Existing Transfer",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("errors"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Transfer reference Id already exist",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&transfer.Transfer{}, nil)
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Error Create Transfer",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("errors"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Failed called ledger ",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				ledgerSvc.On(
					"RecordTransaction",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
				).Return(
					errors.New("error"),
				)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

			},
			expectErr: true,
		},
		{
			name:    "ERROR: Failed update after failed called ledger ",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

				ledgerSvc.On(
					"RecordTransaction",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
				).Return(
					errors.New("error"),
				)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("errors"))
			},
			expectErr: true,
		},
		{
			name:    "ERROR: Failed update after ledger success",
			request: validRequest,
			mockSetup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, platformFeeSvc *mockSvc.IPlatformFeeService, merchantSvc *mockSvc.IMerchantService, mockRepo *mockRepo.ITransferRepository) {
				accountSvc.On(
					"GetMerchantAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
				).Return(
					merchantAccounts,
					nil,
				)
				merchantSvc.On("FindMerchantByID", mock.Anything, recipientId.String()).Return(&merchant.Merchant{}, nil)
				mockRepo.On("GetByReferenceID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("errors"))

				ledgerSvc.On(
					"RecordTransaction",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
				).Return(
					nil,
				)
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

			tc.mockSetup(accSvc, ledgerSvc, platformFeeSvc, merchantSvc, transferRepo)
			svc := New(mockLogger, ledgerSvc, accSvc, platformFeeSvc, merchantSvc, transferRepo)
			response, err := svc.Transfer(context.TODO(), tc.request)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, response)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, response)
				assert.NotEqual(t, uuid.Nil, response.UUID)
			}
		})
	}
}

func TestGetLedgerRequest(t *testing.T) {
	sourceId := uuid.New()
	recipientId := uuid.New()
	parentId := uuid.New()
	id := uuid.New()

	accountMap := map[uuid.UUID]*account_model.Account{
		sourceId:    {UUID: uuid.New()},
		parentId:    {UUID: uuid.New()},
		recipientId: {UUID: uuid.New()},
	}

	testCases := []struct {
		name        string
		request     *transfer.Transfer
		expectErr   bool
		expected    *ledger_model.CreateNewLedgerEntryRequest
		expectedErr error
	}{
		{
			name: "SUCCESS: Get Ledger Request Direct",
			request: &transfer.Transfer{
				UUID:         id,
				MerchantID:   sourceId,
				RecipientID:  recipientId,
				ReferenceID:  "reference-id",
				TransferType: constant.MoneyFlowDirect,
				Amount:       10,
				Remarks:      "remarks",
			},
			expectErr: false,
			expected: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:        id.String(),
				Usecase:            constant.ReferencePlatform,
				TransactionType:    constant.TypeTransfer,
				Channel:            "",
				Remarks:            "remarks",
				Amount:             10,
				Currency:           constant.CurrencyIDR,
				TransferType:       constant.TransferTypeP2P,
				RecipientID:        recipientId,
				RecipientAccountID: accountMap[recipientId].UUID,
				SenderID:           sourceId,
				SenderAccountID:    accountMap[sourceId].UUID,
				ParentID:           parentId,
				ParentAccountID:    accountMap[parentId].UUID,
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
		},
		{
			name: "SUCCESS: Get Ledger Request Indirect",
			request: &transfer.Transfer{
				UUID:         id,
				MerchantID:   sourceId,
				RecipientID:  recipientId,
				ReferenceID:  "reference-id",
				TransferType: constant.MoneyFlowIndirect,
				Amount:       10,
				Remarks:      "remarks",
			},
			expectErr: false,
			expected: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:        id.String(),
				Usecase:            constant.ReferencePlatform,
				TransactionType:    constant.TypeTransfer,
				Channel:            "",
				Remarks:            "remarks",
				Amount:             10,
				Currency:           constant.CurrencyIDR,
				TransferType:       constant.TransferTypeP2P,
				RecipientID:        recipientId,
				RecipientAccountID: accountMap[recipientId].UUID,
				SenderID:           sourceId,
				SenderAccountID:    accountMap[sourceId].UUID,
				ParentID:           parentId,
				ParentAccountID:    accountMap[parentId].UUID,
				MoneyFlowType:      constant.MoneyFlowIndirect,
			},
		},
		{
			name: "ERROR: Empty Source Account",
			request: &transfer.Transfer{
				MerchantID:   uuid.Nil,
				RecipientID:  recipientId,
				ReferenceID:  "reference-id",
				TransferType: constant.MoneyFlowIndirect,
				Amount:       10,
				Remarks:      "remarks",
			},
			expectErr:   true,
			expectedErr: constant.ErrSenderAccountNotFound,
		},
		{
			name: "ERROR: Empty Recipient Account",
			request: &transfer.Transfer{
				MerchantID:   sourceId,
				RecipientID:  uuid.Nil,
				ReferenceID:  "reference-id",
				TransferType: constant.MoneyFlowIndirect,
				Amount:       10,
				Remarks:      "remarks",
			},
			expectErr:   true,
			expectedErr: constant.ErrRecipientAccountNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := New(nil, nil, nil, nil, nil, nil)
			data, err := svc.(*TransferService).getLedgerRequest(context.Background(), tc.request, accountMap, parentId)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, data)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, tc.expected.ReferenceID, data.ReferenceID)
				assert.Equal(t, tc.expected.Usecase, data.Usecase)
				assert.Equal(t, tc.expected.TransactionType, data.TransactionType)
				assert.Equal(t, tc.expected.Channel, data.Channel)
				assert.Equal(t, tc.expected.Remarks, data.Remarks)
				assert.Equal(t, tc.expected.Amount, data.Amount)
				assert.Equal(t, tc.expected.Currency, data.Currency)
				assert.Equal(t, tc.expected.TransferType, data.TransferType)
				assert.Equal(t, tc.expected.RecipientID, data.RecipientID)
				assert.Equal(t, tc.expected.RecipientAccountID, data.RecipientAccountID)
				assert.Equal(t, tc.expected.SenderID, data.SenderID)
				assert.Equal(t, tc.expected.SenderAccountID, data.SenderAccountID)
				assert.Equal(t, tc.expected.ParentID, data.ParentID)
				assert.Equal(t, tc.expected.ParentAccountID, data.ParentAccountID)
				assert.Equal(t, tc.expected.MoneyFlowType, data.MoneyFlowType)
			}
		})
	}
}

func TestUpdateTransferStatus(t *testing.T) {

	repo := mockRepo.NewITransferRepository(t)

	service := &TransferService{
		repo: repo,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Update transfer status",
			setupMock: func() {
				repo.On("Update", constant.ValueCtxMockType(), mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Update transfer status",
			setupMock: func() {
				repo.On("Update", constant.ValueCtxMockType(), mock.Anything).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(
				t, test.wantErr, service.UpdateTransferStatus(
					context.Background(), "123", "321", constant.StatusFailed, util.ValueToPtr(constant.ReasonDescInvalidBeneficiaryAccount),
				),
			)
			repo.AssertExpectations(t)
		})
	}
}
