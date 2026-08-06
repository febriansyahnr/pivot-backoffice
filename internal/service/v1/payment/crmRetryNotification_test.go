package paymentService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentService_CRMRetryNotification(t *testing.T) {
	ctx := context.Background()
	paymentID := "payment-123"
	merchantID := "merchant-456"
	bankReference := "BANK-REF-789"
	processorRefNumber := "PROC-REF-001"
	referenceID := "REF-001"

	testCases := []struct {
		name      string
		payload   *paymentModel.CRMRetryNotificationRequest
		setupMock func(
			paymentRepo *repositoryMocks.IPaymentRepository,
			paymentMethodSvc *serviceMocks.IPaymentMethodService,
			snapCoreRepo *repositoryMocks.ISnapCoreRepository,
		)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - QRIS payment with unified payment method type",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(100000),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodQris,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: referenceID,
					PaymentMethod:     paymentConstant.PAYMENT_METHOD_QRIS,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "100000",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - QRIS payment with legacy payment method type",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(50000),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: referenceID,
					PaymentMethod:     paymentMethod.Type,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "50000",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - non-QRIS payment using processor reference",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(75000),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: processorRefNumber,
					PaymentMethod:     paymentMethod.Type,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "75000",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - payment not found in repository",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, nil)
			},
			wantErr:     true,
			expectedErr: pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found")),
		},
		{
			name: "error - database error when getting payment",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				dbErr := errors.New("database connection error")
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, dbErr)
			},
			wantErr: true,
		},
		{
			name: "error - payment method not found",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentID,
					MerchantID:      merchantID,
					PaymentMethodID: "pm-123",
					Amount:          decimal.NewFromInt(100000),
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(nil, nil)
			},
			wantErr:     true,
			expectedErr: pkgErrors.New(response.HttpErrNotFound, errors.New("payment method not found")),
		},
		{
			name: "error - payment method service error",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentID,
					MerchantID:      merchantID,
					PaymentMethodID: "pm-123",
					Amount:          decimal.NewFromInt(100000),
				}

				svcErr := errors.New("payment method service unavailable")
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(nil, svcErr)
			},
			wantErr: true,
		},
		{
			name: "error - snap core publish payment fails",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(100000),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodQris,
					},
				}

				publishErr := errors.New("snap core service error")
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, mock.Anything).Return(publishErr)
			},
			wantErr: true,
		},
		{
			name: "success - payment with nil processor reference number",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(100000),
					ProcessorReferenceNumber: nil,
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: "",
					PaymentMethod:     paymentMethod.Type,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "100000",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - payment with nil reference ID for unified QRIS",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromInt(100000),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              nil,
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodQris,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: "",
					PaymentMethod:     paymentConstant.PAYMENT_METHOD_QRIS,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "100000",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - payment with decimal amount",
			payload: &paymentModel.CRMRetryNotificationRequest{
				ID:            paymentID,
				BankReference: bankReference,
			},
			setupMock: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodSvc *serviceMocks.IPaymentMethodService,
				snapCoreRepo *repositoryMocks.ISnapCoreRepository,
			) {
				payment := &paymentModel.Payment{
					UUID:                     paymentID,
					MerchantID:               merchantID,
					PaymentMethodID:          "pm-123",
					Amount:                   decimal.NewFromFloat(125000.50),
					ProcessorReferenceNumber: util.ValueToPtr(processorRefNumber),
					ReferenceID:              util.ValueToPtr(referenceID),
				}

				paymentMethod := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodQris,
					},
				}

				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, "pm-123", merchantID).Return(paymentMethod, nil)
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: referenceID,
					PaymentMethod:     paymentConstant.PAYMENT_METHOD_QRIS,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "125000.5",
					},
				}).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)

			tc.setupMock(paymentRepo, paymentMethodSvc, snapCoreRepo)

			service := &PaymentService{
				logger:           logger,
				paymentRepo:      paymentRepo,
				paymentMethodSvc: paymentMethodSvc,
				snapCoreRepo:     snapCoreRepo,
			}

			// Execute
			err := service.CRMRetryNotification(ctx, tc.payload)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			paymentRepo.AssertExpectations(t)
			paymentMethodSvc.AssertExpectations(t)
			snapCoreRepo.AssertExpectations(t)
		})
	}
}
