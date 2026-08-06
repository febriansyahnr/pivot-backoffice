package platformService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantTransactionList(t *testing.T) {
	testCases := []struct {
		name    string
		input   *platform.TransactionRequest
		setup   func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService)
		wantErr bool
	}{
		{
			name: "SUCCESS: DISBURSEMENT Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.TypeDisbursement,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "SUCCESS: PAYMENT Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.TypePayment,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				paymentSvc.On(
					"FilterPaymentHistory",
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "SUCCESS: PLATFORM TRANSFER Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferencePlatformTransfer,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				transferSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "SUCCESS: WITHDRAWAL Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferenceWithdrawal,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				withdrawalSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "SUCCESS: TOP_UP Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferenceTopUp,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				merchantTopUpSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "ERROR: Invalid submerchant",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.TypeDisbursement,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: DISBURSEMENT Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.TypeDisbursement,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: PAYMENT Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.TypePayment,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				paymentSvc.On(
					"FilterPaymentHistory",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: PLATFORM TRANSFER Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferencePlatformTransfer,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				transferSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: WITHDRAWAL Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferenceWithdrawal,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				withdrawalSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: TOP_UP Request",
			input: &platform.TransactionRequest{
				ParentMerchantId: uuid.NewString(),
				MerchantId:       uuid.NewString(),
				Reference:        constant.ReferenceTopUp,
			},
			setup: func(merchantSvc *mocks.IMerchantService, disbursementSvc *mocks.IDisbursementService, paymentSvc *mocks.IPaymentService, transferSvc *mocks.ITransferService, withdrawalSvc *mocks.IWithdrawalService, merchantTopUpSvc *mocks.IMerchantTopUpService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				merchantTopUpSvc.On(
					"GetList",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementSvc := mocks.NewIDisbursementService(t)
			paymentSvc := mocks.NewIPaymentService(t)
			transferSvc := mocks.NewITransferService(t)
			withdrawalSvc := mocks.NewIWithdrawalService(t)
			merchantTopUpSvc := mocks.NewIMerchantTopUpService(t)
			merchantSvc := mocks.NewIMerchantService(t)
			logger, _ := logger.NewZapLogger(logger.Config{})
			tc.setup(merchantSvc, disbursementSvc, paymentSvc, transferSvc, withdrawalSvc, merchantTopUpSvc)

			svc := New(logger, disbursementSvc, paymentSvc, merchantSvc, transferSvc, withdrawalSvc, merchantTopUpSvc)
			_, err := svc.GetMerchantTransactionList(context.Background(), tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
