package merchantTopUp

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindOrCreate(t *testing.T) {
	existedReference := &merchantTopUp.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	existedPayment := &paymentModel.PaymentMethod{
		UUID:      "payment-uuid-uuid",
		Type:      paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  "permata",
		Name:      "permata",
		Acquirer:  "permata",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	validMerchant := &merchantModel.Merchant{
		UUID: "merchant-id",
		Name: "merchant-name",
		MID:  sql.NullString{String: "merchant-mid", Valid: true},
	}

	validSubMerchant := &merchantModel.Merchant{
		UUID:     "sub-merchant-id",
		Name:     "sub-merchant-name",
		ParentID: sql.NullString{String: validMerchant.UUID, Valid: true},
	}

	const VaNumber = "1234123412341234"

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	testCases := []struct {
		name            string
		paymentMethodId string
		mocksSetup      func(reference *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService)
		wantErr         bool
	}{
		{
			name:            "SUCCESS:Merchant top up reference found",
			paymentMethodId: "payment-method-id",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(existedReference, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(existedPayment, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Successfully create merchant top up reference",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(existedPayment, nil)

				snap.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Return(&snapCoreModel.CreateVirtualAccountResponseData{VirtualAccountNo: VaNumber}, nil)

				trxRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "SUCCESS:Successfully create submerchant top up reference",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), mock.Anything,
				).Times(1).Return(validSubMerchant, nil)

				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), validMerchant.UUID,
				).Times(1).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(existedPayment, nil)

				snap.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Return(&snapCoreModel.CreateVirtualAccountResponseData{VirtualAccountNo: VaNumber}, nil)

				trxRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "FAILED:Failed find merchant",
			mocksSetup: func(_ *mocks.IMerchantTopUpRepository, _ *mocks.IPaymentMethodRepository, _ *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "FAILED:Merchant not found",
			mocksSetup: func(_ *mocks.IMerchantTopUpRepository, _ *mocks.IPaymentMethodRepository, _ *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILED:Failed to get merchant top up reference",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, _ *mocks.IPaymentMethodRepository, _ *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "FAILED:Failed to get payment method",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Non virtual account type payment_method",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(&paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_BANK_TRANSFER,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILED:Failed to create virtual_account to snap_core",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(existedPayment, nil)

				snap.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "FAILED: failed to create disbursement_top_up_reference",
			mocksSetup: func(trxRepo *mocks.IMerchantTopUpRepository, payment *mocks.IPaymentMethodRepository, snap *mocks.ISnapCoreRepository, merchantSvc *mockSvc.IMerchantService) {
				merchantSvc.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(validMerchant, nil)

				trxRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId",
					mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, nil)

				payment.On(
					"GetPaymentMethodById", mock.Anything, constant.StringMockType(),
				).Return(existedPayment, nil)

				snap.On(
					"CreateVirtualAccount", mock.Anything, mock.Anything,
				).Return(&snapCoreModel.CreateVirtualAccountResponseData{
					VirtualAccountNo: VaNumber,
				}, nil)

				trxRepo.On("Create", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			cfg := &config.Config{
				ServiceName: "testing",
			}

			refRepo := mocks.NewIMerchantTopUpRepository(t)
			paymentMock := mocks.NewIPaymentMethodRepository(t)
			snapCoreMock := mocks.NewISnapCoreRepository(t)
			merchantSvc := mockSvc.NewIMerchantService(t)

			tc.mocksSetup(refRepo, paymentMock, snapCoreMock, merchantSvc)

			trxSvc := New(cfg, log, paymentMock, refRepo, snapCoreMock, WithMerchantService(merchantSvc))

			reference, err := trxSvc.FindOrCreate(context.Background(), uuid.NewString(), "", tc.paymentMethodId)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, reference)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, reference)
			}

			refRepo.AssertExpectations(t)

			paymentMock.AssertExpectations(t)
			snapCoreMock.AssertExpectations(t)
		})
	}
}
