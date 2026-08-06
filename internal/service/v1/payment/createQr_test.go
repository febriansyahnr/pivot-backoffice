package paymentService

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestCreatePaymentQris(t *testing.T) {
	feeSvc := serviceMocks.NewIFeeService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)

	mockRmq := mockRabbitMq.NewRabbitMQExt(t)
	mockRmq.On(
		"PublishWithDelay",
		mock.Anything,
		rabbitMqExt.PaymentExpirationRoutingKey,
		mock.Anything,
		mock.AnythingOfType("time.Duration"),
	).Return(nil)

	merchantData := &merchant.Merchant{}
	merchantId := "966f1454-252c-4e28-9a99-1803558cd5a5"
	subMerchantId := "dcdf92f2-9f3a-4606-94b1-5d610452095b"
	parentMerchantId := "ab059b2b-17bd-4c0f-b6c8-ee63f6bb5644"
	acquirerMerchantId := util.GenerateULID()
	ctxValue := context.WithValue(context.Background(), constant.CtxTraceIdKey, "3a61b9f5-737d-412c-ab50-2c9da9870f45")

	validPaymentMethod := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			Acquirer: "bnc",
			Type:     paymentConstant.PAYMENT_METHOD_QRIS,
		},
	}

	testCases := []struct {
		desc         string
		ctx          context.Context
		setupPayload func() paymentModel.PaymentRequest
		wantErr      bool
		setupMock    func(
			paymentRepoMocks *repositoryMocks.IPaymentRepository,

			snapCoreMocks *repositoryMocks.ISnapCoreRepository,
			customerRepoMocks *repositoryMocks.ICustomerRepository,
			merchantRepoMocks *repositoryMocks.IMerchantRepository,
			paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
			qrisSvc *serviceMocks.IQrisService,
			feeSvc *serviceMocks.IFeeService,
		)
	}{
		{
			desc:    "success payment qris static",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:       constant.QrTypeStatic,
						QrMethodType: constant.QrMethodTypeMPM,
						Amount:       &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
					},
					Customer: paymentModel.PaymentRequestCustomer{
						Email: "VJ2jK@example.com",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, mock.AnythingOfType("string")).Return(&customerModel.Customer{
					UUID:       uuid.NewString(),
					Email:      "VJ2jK@example.com",
					MerchantID: "mock-id",
				}, nil)
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)

				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("GetPaymentQrStaticByMerchantId", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)

				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success payment qris dynamic",
			ctx:     ctxValue,
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
					Customer: paymentModel.PaymentRequestCustomer{
						Email: "VJ2jK@example.com",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, mock.AnythingOfType("string")).Return(&customerModel.Customer{
					UUID:       uuid.NewString(),
					Email:      "VJ2jK@example.com",
					MerchantID: "mock-id",
				}, nil)
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, merchantId,
				).Times(1).Return(&merchant.Merchant{
					Name:     "merchant",
					UUID:     merchantId,
					ParentID: sql.NullString{Valid: true, String: parentMerchantId},
				}, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, parentMerchantId,
				).Times(1).Return(&merchant.Merchant{UUID: parentMerchantId}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success payment qris dynamic for qr merchant type direct store",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeSubMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success payment qris dynamic for qr merchant type franchisee",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeFranchisee, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success payment qris static existing data",
			wantErr: false,
			ctx:     ctxValue,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						Amount:        &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
						SubMerchantId: subMerchantId,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On(
					"GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Return(validPaymentMethod, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, subMerchantId,
				).Times(1).Return(&merchant.Merchant{
					Name: "merchant",
					UUID: subMerchantId,
					ParentID: sql.NullString{
						String: merchantId, Valid: true,
					},
				}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, merchantId).Times(1).Return(&merchant.Merchant{UUID: merchantId}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(
					&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil,
				)
				metadata := map[string]any{
					"qrType": "STATIC",
					"isSnap": true,
					"snapCore": map[string]any{
						"partnerReferenceNo": "mock-partner-reference-no",
					},
				}
				paymentRepoMocks.On("GetPaymentQrStaticByMerchantId", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.Payment{
					UUID:            "mock-id",
					ReferenceID:     util.ValueToPtr("mock-reference-id"),
					MerchantID:      "mock-merchant-id",
					PaymentMethodID: "mock-payment-method-id",
					Metadata:        &metadata,
				}, nil)
			},
		},
		{
			desc:    "error get active payment method",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:       constant.QrTypeStatic,
						QrMethodType: constant.QrMethodTypeMPM,
						Amount:       &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(nil, errors.New("error get payment method"))
			},
		},
		{
			desc:    "error not found active payment method",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:       constant.QrTypeStatic,
						QrMethodType: constant.QrMethodTypeMPM,
						Amount:       &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(nil, nil)
			},
		},
		{
			desc:    "error get merchant by id",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						Amount:        &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
						SubMerchantId: "sub-merchant-id",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("error get merchant"))
			},
		},
		{
			desc:    "error not found merchant by id",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						Amount:        &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
						SubMerchantId: "sub-merchant-id",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, nil)
			},
		},
		{
			desc:    "error get qr registration by external id",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						Amount:        &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
						SubMerchantId: subMerchantId,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, subMerchantId,
				).Times(1).Return(&merchant.Merchant{
					ParentID: sql.NullString{
						String: merchantId, Valid: true,
					},
					Name: "merchant",
					UUID: subMerchantId,
				}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, merchantId).Times(1).Return(&merchant.Merchant{UUID: merchantId}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error get qr registration by external id merchant not found",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:       constant.QrTypeStatic,
						QrMethodType: constant.QrMethodTypeMPM,
						Amount:       &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, merchantId).Return(&merchant.Merchant{Name: "merchant", UUID: merchantId}, nil)
				qrisSvc.On("FindQrRegistrationByExternalID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound))
			},
		},
		{
			desc:    "error get payment qris static existing data",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						Amount:        &paymentModel.Amount{Value: decimal.NewFromInt(0), Currency: "IDR"},
						SubMerchantId: subMerchantId,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, subMerchantId,
				).Return(&merchant.Merchant{
					UUID: subMerchantId,
					ParentID: sql.NullString{
						String: merchantId, Valid: true,
					},
				}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, merchantId).Return(&merchant.Merchant{UUID: merchantId}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				paymentRepoMocks.On("GetPaymentQrStaticByMerchantId", mock.Anything, mock.Anything, subMerchantId, mock.Anything).Return(nil, errors.New("error get payment qr static"))
			},
		},
		{
			desc:    "error request to snap core",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(nil, errors.New("error when generate qr mpm"))
			},
		},
		{
			desc:    "error begin tx",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(nil, errors.New("error begin tx"))
			},
		},
		{
			desc:    "error when create payment and failed to rollback tx",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(errors.New("error create payment"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).Return(errors.New("error rollback tx"))
			},
		},
		{
			desc:    "error when create payment but success to rollback tx",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(errors.New("error create payment"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error when commit tx",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:         constant.QrTypeDynamic,
						QrMethodType:   constant.QrMethodTypeMPM,
						Amount:         &paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
						ValidityPeriod: 3600,
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,

				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				paymentMethodRepoMocks *repositoryMocks.IPaymentMethodRepository,
				qrisSvc *serviceMocks.IQrisService,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodRepoMocks.On("GetActivePaymentMethodByRequest", mock.Anything, constant.PtrGetPaymentMethodFilterRequestMockType()).
					Return(validPaymentMethod, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, mock.AnythingOfType("string")).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				qrisSvc.On(
					"FindQrRegistrationByExternalID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&qris.Registration{MerchantType: constant.QrMerchantTypeMerchant, AcquirerMerchantId: &acquirerMerchantId}, nil)
				snapCoreMocks.On("GenerateQrMpm", mock.Anything, mock.AnythingOfType("snapCoreModel.GenerateQrMpmRequest")).
					Return(&snapCoreModel.GenerateQrMpmResponseData{
						CreatedAt: time.Now(),
						Amount:    commonModel.Amount{Value: "10000", Currency: "IDR"},
						FeeAmount: commonModel.Amount{Value: "0", Currency: "IDR"},
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).Return(errors.New("error commit transaction"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			paymentRepoMocks := repositoryMocks.NewIPaymentRepository(t)
			loggerMocks, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			snapCoreMocks := repositoryMocks.NewISnapCoreRepository(t)
			customerRepoMocks := repositoryMocks.NewICustomerRepository(t)
			merchantRepoMocks := repositoryMocks.NewIMerchantRepository(t)
			paymentMethodRepoMocks := repositoryMocks.NewIPaymentMethodRepository(t)
			accountRepoMock := repositoryMocks.NewIAccountRepository(t)
			qrisSvc := serviceMocks.NewIQrisService(t)

			tc.setupMock(paymentRepoMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, paymentMethodRepoMocks, qrisSvc, feeSvc)

			paymentSvc := New(paymentRepoMocks, loggerMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, paymentMethodRepoMocks, accountRepoMock,
				WithQrisService(qrisSvc),
				WithConfig(&config.Config{
					Environment: constant.EnvironmentStaging,
					MerchantPortalConfig: config.MerchantPortalConfig{
						PaymentSimulationPatternURL: "https://dashboard-stg.harsya.com/simulation/payment/%s",
					},
					UnifiedPaymentConfig: config.UnifiedPaymentConfig{
						QrConfig: &config.UnifiedPaymentQrConfig{
							MaxExpiryDuration:     60,
							MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
						},
					},
				}),
				WithFeeService(feeSvc),
				WithOrchestratorService(orchestratorSvc),
				WithRabbitMQClient(mockRmq),
			)

			if tc.ctx == nil {
				tc.ctx = context.WithValue(context.Background(), constant.CtxParentMerchantId, parentMerchantId)
			}
			result, err := paymentSvc.CreatePayment(tc.ctx, merchantId, tc.setupPayload())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			paymentRepoMocks.AssertExpectations(t)

			snapCoreMocks.AssertExpectations(t)
			customerRepoMocks.AssertExpectations(t)
			merchantRepoMocks.AssertExpectations(t)
		})
	}
}

func TestPublishQRExpiryMessageWithActualTime(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name          string
		payment       *paymentModel.PaymentDTO
		shouldPublish bool
		shouldError   bool
	}{
		{
			name: "payment expires in 15 minutes - should publish",
			payment: &paymentModel.PaymentDTO{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				ExpiredAt:  util.ValueToPtr(now.Add(15 * time.Minute)),
			},
			shouldPublish: true,
		},
		{
			name: "payment expires at 17:59 UTC - should publish",
			payment: &paymentModel.PaymentDTO{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				ExpiredAt:  util.ValueToPtr(time.Date(now.Year(), now.Month(), now.Day(), 17, 59, 0, 0, time.UTC)),
			},
			shouldPublish: true,
		},
		{
			name: "payment expires at two days later UTC - should not publish",
			payment: &paymentModel.PaymentDTO{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				ExpiredAt:  util.ValueToPtr(time.Date(now.Year(), now.Month(), now.Day()+2, 18, 1, 0, 0, time.UTC)),
			},
		},
		{
			name: "payment expires at 17:59 UTC - should publish but got error",
			payment: &paymentModel.PaymentDTO{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				ExpiredAt:  util.ValueToPtr(time.Date(now.Year(), now.Month(), now.Day(), 17, 59, 0, 0, time.UTC)),
			},
			shouldPublish: false,
			shouldError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			_, pdkLog, _ := test.SetupLogger()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			// Create a new paymentService manually for testing
			paymentService := &PaymentService{
				rabbitMqExt: mockRmq,
				logger:      pdkLog,
			}

			// Set up expectations
			if tc.shouldPublish {
				mockRmq.On(
					"PublishWithDelay",
					mock.Anything,
					rabbitMqExt.PaymentExpirationRoutingKey,
					&paymentModel.ExpiringPayment{
						UUID:       tc.payment.UUID,
						MerchantID: tc.payment.MerchantID,
						ExpiredAt:  *tc.payment.ExpiredAt,
					},
					mock.AnythingOfType("time.Duration"),
				).Return(nil).Once()
			}

			if tc.shouldError {
				mockRmq.On(
					"PublishWithDelay",
					mock.Anything,
					rabbitMqExt.PaymentExpirationRoutingKey,
					&paymentModel.ExpiringPayment{
						UUID:       tc.payment.UUID,
						MerchantID: tc.payment.MerchantID,
						ExpiredAt:  *tc.payment.ExpiredAt,
					},
					mock.AnythingOfType("time.Duration"),
				).Return(constant.ErrSomeErrorForUnitTest).Once()
			}

			// Call the method under test
			paymentService.PublishQRExpiryMessage(context.Background(), tc.payment)

			// Verify expectations
			if tc.shouldPublish {
				mockRmq.AssertExpectations(t)
			} else {
				mockRmq.AssertNotCalled(t, "PublishWithDelay")
			}
		})
	}
}
