package paymentService

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/go/snap/structs/va"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

const VaNumber = "1234123412341234"

func TestCreatePayment(t *testing.T) {
	dataCustomer := paymentModel.PaymentRequestCustomer{
		Name:  "name",
		Email: "name@email.co",
		Phone: "081234567890",
	}
	virtualAccountPayload := paymentModel.PaymentRequest{
		ReferenceID:   uuid.NewString(),
		PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		TotalAmount:   paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
		Customer:      dataCustomer,
		PaymentItems: &[]paymentModel.PaymentItemRequest{
			{
				Name: "nasi goreng",
				Qty:  1,
				Amount: paymentModel.Amount{
					Value:    decimal.NewFromInt(10000),
					Currency: "IDR",
				},
			},
		},
		VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
			VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
			Issuer:                constant.BANK_ACQUIRER_PERMATA,
			VirtualAccountName:    "validName",
		},
	}

	feeSvc := serviceMocks.NewIFeeService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)

	orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)

	merchantId := "f03754a0-0e9b-4955-ae92-32abddd60260"
	parentMerchantId := "e3189d2f-de94-4059-825e-fa6e4204de49"
	ctxValue := context.WithValue(context.Background(), constant.CtxTraceIdKey, "b406a57b-266b-4316-afcc-103c16fc7079")

	validPaymentMethod := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			Acquirer: "permata",
			Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		},
	}

	testCases := []struct {
		ctx          context.Context
		desc         string
		setupPayload func() paymentModel.PaymentRequest
		wantErr      bool
		setupMock    func(
			paymentRepoMocks *repositoryMocks.IPaymentRepository,
			snapCoreMocks *repositoryMocks.ISnapCoreRepository,
			customerRepoMocks *repositoryMocks.ICustomerRepository,
			merchantRepoMocks *repositoryMocks.IMerchantRepository,
			feeSvc *serviceMocks.IFeeService,
		)
	}{
		{
			desc:    "success payment method virtual account",
			ctx:     ctxValue,
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				firstName, lastName := customerModel.FullNameToFirstNameAndLastName(dataCustomer.Name)
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: firstName,
						LastName:  lastName,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, merchantId,
				).Times(1).Return(&merchant.Merchant{
					UUID:     merchantId,
					ParentID: sql.NullString{Valid: true, String: parentMerchantId},
				}, nil)
				merchantRepoMocks.On(
					"FindMerchantByID", mock.Anything, parentMerchantId,
				).Times(1).Return(&merchant.Merchant{UUID: parentMerchantId}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "error method not allowed",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				virtualAccountPayload.PaymentMethod = "not allowed"
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				firstName, lastName := customerModel.FullNameToFirstNameAndLastName(dataCustomer.Name)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, mock.Anything).
					Return(&customerModel.Customer{
						FirstName: firstName,
						LastName:  lastName,
						Email:     dataCustomer.Email,
					}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			paymentRepoMocks := repositoryMocks.NewIPaymentRepository(t)
			snapCoreMocks := repositoryMocks.NewISnapCoreRepository(t)
			customerRepoMocks := repositoryMocks.NewICustomerRepository(t)
			merchantRepoMocks := repositoryMocks.NewIMerchantRepository(t)
			tc.setupMock(paymentRepoMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, feeSvc)

			paymentSvc := New(paymentRepoMocks, nil, snapCoreMocks, customerRepoMocks, merchantRepoMocks, nil, nil,
				WithConfig(&config.Config{
					Environment: constant.EnvironmentStaging,
					MerchantPortalConfig: config.MerchantPortalConfig{
						PaymentSimulationPatternURL: "https://dashboard-stg.harsya.com/simulation/payment/%s",
					},
				}),
				WithFeeService(feeSvc),
				WithOrchestratorService(orchestratorSvc),
				WithPaymentMethodService(paymentMethodSvc),
			)
			if tc.ctx == nil {
				tc.ctx = context.WithValue(context.Background(), constant.CtxParentMerchantId, uuid.NewString())
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

func TestIntegrationCreatePaymentUsingVirtualAccount(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	dataCustomer := paymentModel.PaymentRequestCustomer{
		Name:  "name",
		Email: "name@email.co",
		Phone: "081234567890",
	}

	ctx := context.Background()

	logger, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)
	consulContainer, consulURL, err := test.SetupConsul(ctx)
	assert.NoError(t, err)
	test.SetupFeatureFlag(consulURL)
	test.SetupGoff(ctx, consulURL, pdkLogger)

	feeSvc := serviceMocks.NewIFeeService(t)

	virtualAccountPayload := paymentModel.PaymentRequest{
		ReferenceID:   uuid.NewString(),
		PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		TotalAmount:   paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
		Customer:      dataCustomer,
		PaymentItems: &[]paymentModel.PaymentItemRequest{
			{
				Name: "nasi goreng",
				Qty:  1,
				Amount: paymentModel.Amount{
					Value:    decimal.NewFromInt(10000),
					Currency: "IDR",
				},
				Metadata:    &map[string]any{},
				Description: "DESCRIPTION",
			},
		},
		VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
			VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
			Issuer:                constant.BANK_ACQUIRER_PERMATA,
			VirtualAccountName:    "validName",
		},
	}
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)

	merchantData := &merchant.Merchant{}

	validPaymentMethod := &paymentModel.PaymentMethodWithPivot{
		PaymentMethod: paymentModel.PaymentMethod{
			Acquirer: "permata",
			Type:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		},
	}

	testCases := []struct {
		desc            string
		setupPayload    func() paymentModel.PaymentRequest
		wantErr         bool
		consulRetriever bool
		setupMock       func(
			paymentRepoMocks *repositoryMocks.IPaymentRepository,

			snapCoreMocks *repositoryMocks.ISnapCoreRepository,
			customerRepoMocks *repositoryMocks.ICustomerRepository,
			merchantRepoMocks *repositoryMocks.IMerchantRepository,
			feeSvc *serviceMocks.IFeeService,

		)
	}{
		{
			desc:    "error amount reach maximum amount in payment",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				invalidAcqAndType := virtualAccountPayload
				invalidAcqAndType.TotalAmount.Value = decimal.NewFromInt(250000001)
				return invalidAcqAndType
			},
			consulRetriever: true,
			setupMock: func(_ *repositoryMocks.IPaymentRepository, _ *repositoryMocks.ISnapCoreRepository, _ *repositoryMocks.ICustomerRepository, _ *repositoryMocks.IMerchantRepository, _ *serviceMocks.IFeeService) {
				/* NOSONAR */
			},
		},
		{
			desc:    "SUCCESS: Create submerchant payment using virtual account",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name:     "merchant",
						UUID:     "mock-id",
						ParentID: sql.NullString{String: "mock-parent-id"},
					}, nil)

				merchantRepoMocks.On("FindMerchantByID", mock.Anything, "mock-parent-id").
					Return(&merchant.Merchant{
						Name: "parent-merchant",
						UUID: "mock-parent-id",
						MID:  sql.NullString{String: "mock-mid"},
					}, nil)

				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "error get payment active payment method",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				invalidAcqAndType := virtualAccountPayload
				invalidAcqAndType.PaymentMethod = "invalid_method"
				return invalidAcqAndType
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(nil, errors.New("error get payment method"))
			},
		},
		{
			desc:    "get payment active payment method not found",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				invalidAcqAndType := virtualAccountPayload
				invalidAcqAndType.PaymentMethod = "invalid_method"
				return invalidAcqAndType
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(nil, nil)
			},
		},
		{
			desc:    "split route config do not apply to facilitator payment methods",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				request := virtualAccountPayload
				request.SplitRoutingConfigurations = &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
					{
						MerchantId:  "73c39ae0-a2a9-4edf-84b5-3405c6832f93",
						Type:        "FIXED", // NOSONAR
						Currency:    "IDR",   // NOSONAR
						FixedAmount: 1_000,   // NOSONAR
					},
				}
				return request
			},
			setupMock: func(
				_ *repositoryMocks.IPaymentRepository,
				_ *repositoryMocks.ISnapCoreRepository,
				_ *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				_ *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(&paymentModel.PaymentMethodWithPivot{
						ChannelType: constant.PaymentMethodChannelTypeDirect,
					}, nil)
			},
		},
		{
			desc:    "error database when find customer by email",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)

				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(nil, errors.New("error when find customer by email"))
			},
		},
		{
			desc:    "error database when create data customer",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(merchantData, nil)

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(nil, nil)

				customerRepoMocks.On("Create", mock.Anything, mock.Anything).Return(errors.New("error when create data customer"))
			},
		},
		{
			desc:    "error FindMerchantByID",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			desc:    "error merchant not found",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).Return(nil, nil)
			},
		},
		{
			desc:    "error merchant found but error do http to snapCore",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				virtualAccountPayload.VirtualAccount.VirtualAccountTrxType = paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC
				virtualAccountPayload.VirtualAccount.MaxAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(100000),
					Currency: "IDR",
				}
				virtualAccountPayload.VirtualAccount.MinAmount = &paymentModel.Amount{
					Value:    decimal.NewFromInt(10000),
					Currency: "IDR",
				}
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).Return(nil, errors.New("error snap core"))
			},
		},
		{
			desc:    "success with minAmount is not nil",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount:   paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
					Customer:      dataCustomer,
					PaymentItems: &[]paymentModel.PaymentItemRequest{
						{
							Name: "nasi goreng",
							Qty:  1,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(10000),
								Currency: "IDR",
							},
						},
					},
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(100000),
							Currency: "IDR",
						},
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(10000),
							Currency: "IDR",
						},
						Issuer:             constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountName: "validName",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						MinAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success with maxAmount is not nil",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					TotalAmount:   paymentModel.Amount{Value: decimal.NewFromInt(10000), Currency: "IDR"},
					Customer:      dataCustomer,
					PaymentItems: &[]paymentModel.PaymentItemRequest{
						{
							Name: "nasi goreng",
							Qty:  1,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(10000),
								Currency: "IDR",
							},
						},
					},
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
						MinAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(10000),
							Currency: "IDR",
						},
						MaxAmount: &paymentModel.Amount{
							Value:    decimal.NewFromInt(100000),
							Currency: "IDR",
						},
						Issuer:             constant.BANK_ACQUIRER_PERMATA,
						VirtualAccountName: "validName",
					},
				}
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						MinAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						MaxAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "error begin transaction database",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)

				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(nil, errors.New("error in begin transaction"))
			},
		},
		{
			desc:    "error database create payment err rollback transaction",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)

				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(errors.New("error when create payment"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).
					Return(errors.New("error in rollback transaction"))
			},
		},
		{
			desc:    "error database create payment success rollback transaction",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)

				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(errors.New("error when create payment"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error database create payment items err rollback transaction",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)

				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(errors.New("error when create payment item"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).
					Return(errors.New("error when rollback transaction"))
			},
		},
		{
			desc:    "error database create payment items success rollback transaction",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)

				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(errors.New("error when create payment item"))
				paymentRepoMocks.On("RollbackTransaction", mock.Anything).Return(nil)
			},
		},
		{
			desc:    "error commit transaction",
			wantErr: true,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)

				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(errors.New("error commit transaction"))
			},
		},
		{
			desc:    "success create payment using virtual account with empty account name",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				virtualAccountPayload.VirtualAccount.VirtualAccountName = ""
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success create payment using virtual account",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
		},
		{
			desc:    "success create payment with bill Detail",
			wantErr: false,
			setupPayload: func() paymentModel.PaymentRequest {
				virtualAccountPayload.VirtualAccount.BillDetails = &[]va.BillDetail{
					{
						BillName: "sample bill name",
						BillAmount: va.Amount{
							Value:    "10000",
							Currency: "IDR",
						},
					},
				}
				return virtualAccountPayload
			},
			setupMock: func(
				paymentRepoMocks *repositoryMocks.IPaymentRepository,
				snapCoreMocks *repositoryMocks.ISnapCoreRepository,
				customerRepoMocks *repositoryMocks.ICustomerRepository,
				merchantRepoMocks *repositoryMocks.IMerchantRepository,
				feeSvc *serviceMocks.IFeeService,
			) {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything).Once().
					Return(validPaymentMethod, nil)
				customerRepoMocks.On("FindCustomerByEmail", mock.Anything, constant.StringMockType()).
					Return(&customerModel.Customer{
						FirstName: dataCustomer.Name,
						Email:     dataCustomer.Email,
					}, nil)
				merchantRepoMocks.On("FindMerchantByID", mock.Anything, constant.StringMockType()).
					Return(&merchant.Merchant{
						Name: "merchant",
						UUID: "mock-id",
					}, nil)
				snapCoreMocks.On("CreateVirtualAccount", mock.Anything, mock.AnythingOfType("snapCoreModel.CreateVirtualAccountRequest")).
					Return(&snapCoreModel.CreateVirtualAccountResponseData{
						ID:               "mock-id",
						VirtualAccountNo: "mock-virtual-account-no",
						AccountName:      "mock-account-name",
						Acquirer:         "mock-acquirer",
						ExpiredAt:        time.Now().Add(time.Hour * 24),
						Status:           paymentConstant.PAYMENT_STATUS_PENDING,
						TotalAmount: snapCoreModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CreatedAt: time.Now(),
					}, nil)
				paymentRepoMocks.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				paymentRepoMocks.On("CreatePayment", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).
					Return(nil)
				paymentRepoMocks.On("CreatePaymentItem", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentItemDTO")).
					Return(nil)
				paymentRepoMocks.On("CommitTransaction", mock.Anything).
					Return(nil)
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

			tc.setupMock(paymentRepoMocks, snapCoreMocks, customerRepoMocks, merchantRepoMocks, feeSvc)

			paymentSvc := PaymentService{
				paymentRepo:  paymentRepoMocks,
				logger:       loggerMocks,
				snapCoreRepo: snapCoreMocks,
				customerRepo: customerRepoMocks,
				merchantRepo: merchantRepoMocks,
				config: &config.Config{
					Environment: constant.EnvironmentStaging,
					MerchantPortalConfig: config.MerchantPortalConfig{
						PaymentSimulationPatternURL: "https://dashboard-stg.harsya.com/simulation/payment/%s",
					},
				},
				feeSvc:           feeSvc,
				orchestratorSvc:  orchestratorSvc,
				paymentMethodSvc: paymentMethodSvc,
			}
			result, err := paymentSvc.createPaymentUsingVirtualAccount(ctx, "merchantID", tc.setupPayload())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if tc.consulRetriever {
				defer pdkLogger.Sync()
				defer logger.Sync()
				defer ffclient.Close()
				defer consulContainer.Terminate(ctx)
			}

			paymentRepoMocks.AssertExpectations(t)
			snapCoreMocks.AssertExpectations(t)
			customerRepoMocks.AssertExpectations(t)
			merchantRepoMocks.AssertExpectations(t)
			paymentMethodRepoMocks.AssertExpectations(t)
			paymentMethodSvc.AssertExpectations(t)
		})
	}
}
