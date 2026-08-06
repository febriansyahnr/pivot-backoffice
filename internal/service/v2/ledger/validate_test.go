package ledgerService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateTransaction(t *testing.T) {
	senderID := uuid.New()
	senderAccountID := uuid.New()
	parentID := uuid.New()
	parentAccountID := uuid.New()
	recipientID := uuid.New()
	recipientAccountID := uuid.New()

	testCases := []struct {
		Name      string
		Request   ledger_model.CreateNewLedgerEntryRequest
		MockSetup func(
			accRepo *mockRepo.IAccountRepository,
			validatorSvc *mockSvc.ILedgerValidatorService,
			customerSvc *mockSvc.ICustomerService,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Requests with sender, recipient and parent accounts",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					senderAccountID,
				).Return(
					&account_model.Account{UUID: senderAccountID, ReferenceID: senderID, Name: constant.ReferenceDisbursement},
					nil,
				)

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					&account_model.Account{UUID: parentAccountID, ReferenceID: parentID, Name: constant.ReferenceDisbursement},
					nil,
				)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Requests with sender",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					senderAccountID,
				).Return(
					&account_model.Account{UUID: senderAccountID, ReferenceID: senderID, Name: constant.ReferenceDisbursement},
					nil,
				)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Requests from customer sender",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					senderAccountID,
				).Return(
					&account_model.Account{UUID: senderAccountID, ReferenceID: senderID, Name: constant.ReferenceDisbursement, UserType: constant.UserTypeCustomer},
					nil,
				)

				customerSvc.On("FindCustomerByID", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.New().String()}, nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Requests with recipient",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					recipientAccountID,
				).Return(
					&account_model.Account{UUID: recipientID, ReferenceID: senderID, Name: constant.ReferenceDisbursement},
					nil,
				)

			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Requests for customer recipient",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					recipientAccountID,
				).Return(
					&account_model.Account{UUID: recipientID, ReferenceID: senderID, Name: constant.ReferenceDisbursement, UserType: constant.UserTypeCustomer},
					nil,
				)
				customerSvc.On("FindCustomerByID", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.New().String()}, nil)

			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Requests for customer recipient with fee",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
				Fee: ledger_model.FeeRequest{
					RecipientAccountID: parentAccountID,
					Amount:             10,
				},
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					recipientAccountID,
				).Return(
					&account_model.Account{UUID: recipientID, ReferenceID: senderID, Name: constant.ReferenceDisbursement, UserType: constant.UserTypeCustomer},
					nil,
				)

				customerSvc.On("FindCustomerByID", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.New().String()}, nil)

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					&account_model.Account{UUID: recipientID, ReferenceID: senderID, Name: constant.ReferenceDisbursement, UserType: constant.UserTypeMerchant},
					nil,
				)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Incorrect request",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut + "s",
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Error get Recipient account",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					recipientAccountID,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Recipient account not found",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					recipientAccountID,
				).Return(
					nil,
					nil,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Error get sender account",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					senderAccountID,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Sender account not found",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					senderAccountID,
				).Return(
					nil,
					nil,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Error get Parent account",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Parent account not found",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					nil,
					nil,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Get Fee recipient account",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: ledger_model.FeeRequest{
					Amount:             10,
					RecipientAccountID: parentAccountID,
				},
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					nil,
					assert.AnError,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Fee recipient account not found",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: ledger_model.FeeRequest{
					Amount:             10,
					RecipientAccountID: parentAccountID,
				},
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					nil,
					nil,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Fee recipient account is customer",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: ledger_model.FeeRequest{
					Amount:             10,
					RecipientAccountID: parentAccountID,
				},
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {

				accRepo.On(
					"GetByUUID",
					constant.BackgroundCtxMockType(),
					parentAccountID,
				).Return(
					&account_model.Account{UUID: recipientID, ReferenceID: senderID, Name: constant.ReferenceDisbursement, UserType: constant.UserTypeCustomer},
					nil,
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Empty account list",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				MoneyFlowType:        constant.MoneyFlowIndirect,
			},
			MockSetup: func(accRepo *mockRepo.IAccountRepository, validatorSvc *mockSvc.ILedgerValidatorService, customerSvc *mockSvc.ICustomerService) {
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			accRepo := mockRepo.NewIAccountRepository(t)
			validatorSvc := mockSvc.NewILedgerValidatorService(t)
			customerSvc := mockSvc.NewICustomerService(t)

			tc.MockSetup(accRepo, validatorSvc, customerSvc)

			svc := New(nil, nil, accRepo, nil, customerSvc, nil)
			WithValidatorService(svc, validatorSvc)
			err := svc.ValidateTransaction(context.Background(), uuid.NewString(), &tc.Request)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}

}
