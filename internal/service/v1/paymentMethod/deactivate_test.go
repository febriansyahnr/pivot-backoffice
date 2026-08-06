package paymentMethodService

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestDeactivate(t *testing.T) {
	repo := repositoryMocks.NewIPaymentMethodRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	qrisSvc := serviceMocks.NewIQrisService(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)

	validPaymentMethodID := uuid.NewString()
	validMerchantID := uuid.NewString()
	validMerchantExternalID := util.GenerateULID()

	paymentMethodVa := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID: validPaymentMethodID,
			Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		},
		MerchantID: validMerchantID,
	}

	paymentMethodQris := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			UUID:     validPaymentMethodID,
			Type:     paymentConstant.PAYMENT_METHOD_QRIS,
			Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
		},
		MerchantID: validMerchantID,
		IsActive:   true,
	}

	tests := []struct {
		name         string
		modifierMock func()
		shouldErr    bool
		wantErr      error
	}{
		{
			name: "ERROR:Find merchant by id",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			shouldErr: true,
			wantErr:   assert.AnError,
		},
		{
			name: "ERROR:Merchant not found",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
			},
			shouldErr: true,
			wantErr:   pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Non KYC merchant status",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(&merchantModel.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{String: constant.MerchantKYCTypeNonKYC}}, nil)
			},
			shouldErr: true,
			wantErr:   pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC),
		},
		{
			name: "ERROR:FindPaymentMethodByIdAndMerchant error",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil).Once()

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(nil, assert.AnError)
			},
			shouldErr: true,
			wantErr:   pkgErrors.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:UpsertPaymentMethodMerchantByIdAndMerchant error for VA method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID}, nil).Once()

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodVa, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), paymentMethodVa,
				).Once().Return(assert.AnError)
			},
			shouldErr: true,
			wantErr:   pkgErrors.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:DeActivate for VA method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: uuid.NewString(), ExternalId: validMerchantExternalID}, nil).Once()

				paymentMethodVa.IsActive = false
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodVa, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS:DeActivate for QRIS method type with sync",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil).Once()

				paymentMethodQris.IsActive = true
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodQris, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)

				// Mock QRIS registration lookup by acquirer
				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(&qris.Registration{
					Id:                 "test-reg-id",
					Status:             constant.QrRegistrationStatusSuccess,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					AcquirerMerchantId: util.ValueToPtr("test-merchant-id"),
					AcquirerTerminalId: util.ValueToPtr("test-terminal-id"),
				}, nil)

				// Mock merchant lookup for sync
				merchantRepo.On(
					"FindMerchantForQrRegistration", constant.ValueCtxMockType(), validMerchantID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(&merchantModel.QrisMerchant{
					MCC:       "5699",
					ShortName: "Test Merchant",
				}, nil)

				// Mock sync to SnapCore
				snapCoreRepo.On(
					"QrSyncRegistration", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS:DeActivate for QRIS method type even when sync fails",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil).Once()

				paymentMethodQris.IsActive = true
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodQris, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)

				// Mock QRIS registration lookup by acquirer
				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(&qris.Registration{
					Id:                 "test-reg-id",
					Status:             constant.QrRegistrationStatusSuccess,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					AcquirerMerchantId: util.ValueToPtr("test-merchant-id"),
					AcquirerTerminalId: util.ValueToPtr("test-terminal-id"),
				}, nil)

				// Mock merchant lookup for sync
				merchantRepo.On(
					"FindMerchantForQrRegistration", constant.ValueCtxMockType(), validMerchantID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(&merchantModel.QrisMerchant{
					MCC:       "5699",
					ShortName: "Test Merchant",
				}, nil)

				// Mock sync to SnapCore - failure should not fail deactivation
				snapCoreRepo.On(
					"QrSyncRegistration", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(assert.AnError)
			},
		},
		{
			name: "SUCCESS:DeActivate for QRIS method type when QR registration not found",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil).Once()

				paymentMethodQris.IsActive = true
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodQris, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)

				// Mock QRIS registration lookup failure - should not fail deactivation
				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(nil, assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modifierMock()

			svc := New(logger, repo, snapCoreRepo, creditCardRepo, WithMerchantService(merchantSvc), WithQrisService(qrisSvc), WithMerchantRepository(merchantRepo))
			err := svc.Deactivate(context.Background(), &paymentModel.PaymentMethodWithPivot{MerchantID: validMerchantID, PaymentMethod: paymentModel.PaymentMethod{UUID: validPaymentMethodID}})

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
