package platformFeeService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplyMerchantTransferFee(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Apply Merchant Fee",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Record to ledger",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("errors"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := logger.NewZapLogger(logger.Config{})
			accountSvc := mockSvc.NewIAccountService(t)
			ledgerSvc := mockSvc.NewILedgerService(t)
			feeSvc := mockSvc.NewIFeeService(t)
			tc.setup(accountSvc, ledgerSvc, feeSvc)

			svc := New(log, ledgerSvc, feeSvc, accountSvc)
			err := svc.(*PlatformFeeService).ApplyMerchantTransferFee(context.Background(), platformFee.PlatformFeeRequest{
				MerchantID:  uuid.NewString(),
				ReferenceID: uuid.NewString(),
				Amount:      1999,
			})
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestApplyMerchantTransactionFee(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Apply Merchant Fee",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Record to ledger",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("errors"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := logger.NewZapLogger(logger.Config{})
			accountSvc := mockSvc.NewIAccountService(t)
			ledgerSvc := mockSvc.NewILedgerService(t)
			feeSvc := mockSvc.NewIFeeService(t)
			tc.setup(accountSvc, ledgerSvc, feeSvc)

			svc := New(log, ledgerSvc, feeSvc, accountSvc)
			err := svc.(*PlatformFeeService).ApplyMerchantTransactionFee(context.Background(), platformFee.PlatformFeeRequest{
				MerchantID:  uuid.NewString(),
				ReferenceID: uuid.NewString(),
				Amount:      1999,
			})
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestApplyMerchantFee(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Apply Merchant Fee",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Apply Merchant Fee, Fee 0",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On(
					"GetFeeCalculationAndDetail",
					mock.Anything,
					mock.Anything,
				).Return(float64(0), &feeModel.FeeMetadataObject{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get Fee",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get Account",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					errors.New("errors"),
				)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Record to ledger",
			setup: func(accountSvc *mockSvc.IAccountService, ledgerSvc *mockSvc.ILedgerService, feeSvc *mockSvc.IFeeService) {
				feeSvc.On("GetFeeCalculationAndDetail", mock.Anything, mock.Anything).Return(1_000.0, &feeModel.FeeMetadataObject{}, nil)
				accountSvc.On(
					"GetAccountByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&account_model.Account{},
					nil,
				)

				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("errors"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := logger.NewZapLogger(logger.Config{})
			accountSvc := mockSvc.NewIAccountService(t)
			ledgerSvc := mockSvc.NewILedgerService(t)
			feeSvc := mockSvc.NewIFeeService(t)
			tc.setup(accountSvc, ledgerSvc, feeSvc)

			svc := New(log, ledgerSvc, feeSvc, accountSvc)
			err := svc.(*PlatformFeeService).ApplyMerchantFee(context.Background(), platformFee.PlatformFeeRequest{
				MerchantID:  uuid.NewString(),
				ReferenceID: uuid.NewString(),
				Amount:      1999,
			}, "referenceType")
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
