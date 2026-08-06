package paymentMethodService

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivate(t *testing.T) {
	repo := repositoryMocks.NewIPaymentMethodRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	qrisSvc := serviceMocks.NewIQrisService(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

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
	}

	tests := []struct {
		name         string
		modifierMock func()
		wantErr      error
	}{
		{
			name: "ERROR:Find merchant by id",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "ERROR:Merchant not found",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR:Non KYC merchant status",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(&merchantModel.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{String: constant.MerchantKYCTypeNonKYC}}, nil)
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC),
		},
		{
			name: "ERROR:FindPaymentMethodByIdAndMerchant error",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:Active payment method",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(&paymentModel.PaymentMethodWithPivot{IsActive: true}, nil)
			},
			wantErr: pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodAlreadyActive),
		},
		{
			name: "ERROR:UpsertPaymentMethodMerchantByIdAndMerchant error for VA method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodVa, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), paymentMethodVa,
				).Once().Return(assert.AnError)
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:Activate for VA method type",
			modifierMock: func() {
				paymentMethodVa.IsActive = false

				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Once().Return(paymentMethodVa, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:FindQrRegistrationByExternalIDAndAcquirer error for QRIS method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Return(paymentMethodQris, nil)

				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:FindQrRegistrationByExternalIDAndAcquirer status not SUCCESS for QRIS method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Return(paymentMethodQris, nil)

				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Once().Return(&qris.Registration{
					Status:   "Not Success",
					Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				}, nil)
			},
			wantErr: pkgErrors.New(response.HttpErrRequest, constant.ErrQrRegistrationIsNotCompleted),
		},
		{
			name: "ERROR:UpsertPaymentMethodMerchantByIdAndMerchant error for QRIS method type",
			modifierMock: func() {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Return(paymentMethodQris, nil)

				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Return(&qris.Registration{
					Status:   constant.QrRegistrationStatusSuccess,
					Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				}, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:Activate for QRIS method type",
			modifierMock: func() {
				paymentMethodQris.IsActive = false

				// Mock FindMerchantByID
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				// Mock FindPaymentMethodByIdAndMerchant to return QRIS payment method
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Return(paymentMethodQris, nil)

				// Mock FindQrRegistrationByExternalIDAndAcquirer
				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Return(&qris.Registration{
					Id:                 "test-reg-id",
					Status:             constant.QrRegistrationStatusSuccess,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					AcquirerMerchantId: util.ValueToPtr("test-merchant-id"),
					AcquirerTerminalId: util.ValueToPtr("test-terminal-id"),
				}, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), paymentMethodQris,
				).Once().Return(nil)

				// Mock merchant lookup for sync
				merchantRepo.On(
					"FindMerchantForQrRegistration", constant.ValueCtxMockType(), validMerchantID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Maybe().Return(&merchantModel.QrisMerchant{
					MCC:       "5699",
					ShortName: "Test Merchant",
				}, nil)

				// Mock sync to SnapCore (should not fail activation even if sync fails)
				snapCoreRepo.On(
					"QrSyncRegistration", constant.ValueCtxMockType(), mock.Anything,
				).Maybe().Return(nil)
			},
		},
		{
			name: "SUCCESS:Activate for QRIS method type even when sync fails",
			modifierMock: func() {
				paymentMethodQris.IsActive = false

				// Mock FindMerchantByID
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: validMerchantID, ExternalId: validMerchantExternalID}, nil)

				// Mock FindPaymentMethodByIdAndMerchant to return QRIS payment method
				repo.On(
					"FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), validPaymentMethodID, validMerchantID,
				).Return(paymentMethodQris, nil)

				// Mock FindQrRegistrationByExternalIDAndAcquirer
				qrisSvc.On(
					"FindQrRegistrationByExternalIDAndAcquirer", constant.ValueCtxMockType(), validMerchantExternalID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Return(&qris.Registration{
					Id:                 "test-reg-id",
					Status:             constant.QrRegistrationStatusSuccess,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					AcquirerMerchantId: util.ValueToPtr("test-merchant-id"),
					AcquirerTerminalId: util.ValueToPtr("test-terminal-id"),
				}, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), paymentMethodQris,
				).Once().Return(nil)

				// Mock merchant lookup for sync
				merchantRepo.On(
					"FindMerchantForQrRegistration", constant.ValueCtxMockType(), validMerchantID, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				).Maybe().Return(&merchantModel.QrisMerchant{
					MCC:       "5699",
					ShortName: "Test Merchant",
				}, nil)

				// Mock sync to SnapCore failure - should not fail activation
				snapCoreRepo.On(
					"QrSyncRegistration", constant.ValueCtxMockType(), mock.Anything,
				).Maybe().Return(assert.AnError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modifierMock()

			svc := New(logger, repo, snapCoreRepo, creditCardRepo, WithQrisService(qrisSvc), WithMerchantService(merchantSvc), WithMerchantRepository(merchantRepo))
			err := svc.Activate(context.Background(), &paymentModel.PaymentMethodWithPivot{MerchantID: validMerchantID, PaymentMethod: paymentModel.PaymentMethod{UUID: validPaymentMethodID}})

			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestActivateInstallment(t *testing.T) {

	defaultInstallmentPaymentMethod := &paymentModel.PaymentMethodWithPivot{
		MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
			PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
					InstallmentPlanIDs: []string{uuid.NewString()},
				},
			},
		},
		PaymentMethod: paymentModel.PaymentMethod{
			Subtype: constant.InstallmentPlanPaymentMethodCard,
			Type:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
		},
	}

	tests := []struct {
		name                     string
		installmentPaymentMethod *paymentModel.PaymentMethodWithPivot
		modifierMock             func(repo *repositoryMocks.IPaymentMethodRepository)
		wantErr                  error
	}{
		{
			name:                     "SUCCESS: Activate installment payment method",
			installmentPaymentMethod: defaultInstallmentPaymentMethod,
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository) {
				repo.On(
					"GetListPaymentMethodMerchant", mock.Anything, mock.Anything,
				).Return(
					[]*paymentModel.PaymentMethodWithPivot{
						{
							PaymentMethod: paymentModel.PaymentMethod{
								UUID: uuid.NewString(),
								Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							},
							IsActive: true,
						},
					}, nil,
				)

				repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:                     "ERROR: Get Card Payment Method",
			installmentPaymentMethod: defaultInstallmentPaymentMethod,
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository) {
				repo.On(
					"GetListPaymentMethodMerchant", mock.Anything, mock.Anything,
				).Return(
					nil, errors.New("error"),
				)

			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errors.New("error")),
		},
		{
			name:                     "ERROR: Empty Active Card Payment Method",
			installmentPaymentMethod: defaultInstallmentPaymentMethod,
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository) {
				repo.On(
					"GetListPaymentMethodMerchant", mock.Anything, mock.Anything,
				).Return(
					[]*paymentModel.PaymentMethodWithPivot{}, nil,
				)

			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrDependentCardPaymentMethodNotActive),
		},
		{
			name: "ERROR: Empty partner config installment",
			installmentPaymentMethod: &paymentModel.PaymentMethodWithPivot{
				MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{},
				},
			},
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository) {
			},
			wantErr: pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrCardInstallmentNotConfigured),
		},
		{
			name:                     "ERROR: Upsert",
			installmentPaymentMethod: defaultInstallmentPaymentMethod,
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository) {
				repo.On(
					"GetListPaymentMethodMerchant", mock.Anything, mock.Anything,
				).Return(
					[]*paymentModel.PaymentMethodWithPivot{
						{
							PaymentMethod: paymentModel.PaymentMethod{
								UUID: uuid.NewString(),
								Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
							},
							IsActive: true,
						},
					}, nil,
				)

				repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			wantErr: pkgErrors.New(response.HttpErrDatabase, errors.New("error")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repositoryMocks.NewIPaymentMethodRepository(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			qrisSvc := serviceMocks.NewIQrisService(t)
			merchantSvc := serviceMocks.NewIMerchantService(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tt.modifierMock(repo)

			merchant := &merchant.Merchant{
				UUID: uuid.NewString(),
				KYCStatus: sql.NullString{
					String: constant.KYCStatusApproved,
					Valid:  true,
				},
			}
			svc := New(logger, repo, snapCoreRepo, creditCardRepo, WithQrisService(qrisSvc), WithMerchantService(merchantSvc))
			err := svc.(*PaymentMethodService).activateInstallment(context.Background(), merchant, tt.installmentPaymentMethod)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
