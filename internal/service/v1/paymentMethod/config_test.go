package paymentMethodService_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	qrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	snapCoreVaModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	paymentMethodService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/paymentMethod"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
)

func TestSetupConfig(t *testing.T) {

	validPaymentMethodID := uuid.NewString()
	validMerchantID := uuid.NewString()
	tests := []struct {
		name         string
		modifierMock func(*repositoryMocks.IPaymentMethodRepository, *serviceMocks.IMerchantService, *repositoryMocks.ISnapCoreRepository, *serviceMocks.IQrisService, *repositoryMocks.IMerchantRepository, *repositoryMocks.ICreditcardCoreProcessorRepository)
		wantErr      bool
		payload      *paymentMethodModel.SetupPaymentMethodConfigRequest
	}{
		{
			name: "ERROR: FindMerchantByID error",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "ERROR: FindMerchantByID not found data",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "ERROR: Merchant KYC status not approved",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Once().Return(&merchantModel.Merchant{ParentID: sql.NullString{Valid: true}, KYCStatus: sql.NullString{Valid: true, String: "NOT REQUIRED"}}, nil)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "ERROR: FindPaymentMethodByIdAndMerchant error",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Return(&merchantModel.Merchant{}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "ERROR: FindPaymentMethodByIdAndMerchant not found data",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Return(&merchantModel.Merchant{}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "ERROR: UpsertPaymentMethodMerchantByIdAndMerchant error",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Return(&merchantModel.Merchant{}, nil)

				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_DANA,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant",
					mock.Anything,
					mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
						SubMerchantID: "sub-merchant-id",
					},
				},
			},
		},
		{
			name: "SUCCESS: Virtual Account",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID",
					mock.Anything,
					mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", MID: sql.NullString{String: "MID123", Valid: true}}, nil)

				repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Processor: c.SnapCoreProcessor,
						Acquirer:  "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				snapCoreRepo.On(
					"GetVirtualAccountConfig",
					mock.Anything,
					mock.Anything,
				).Return([]*snapCoreVaModel.VirtualAccountConfigResponseData{}, nil)

				snapCoreRepo.On(
					"CreateVirtualAccountConfig",
					mock.Anything,
					mock.Anything,
				).Return(&snapCoreVaModel.VirtualAccountConfigResponseData{}, nil)

				snapCoreRepo.On(
					"UpdateVirtualAccountConfigPrefix",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				ChannelType:     "API",
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
							{
								BINPrefix:  "123",
								Type:       "OPEN_STATIC",
								StartRange: "0001",
								EndRange:   "9999",
							},
						},
					},
				},
			},
		},
		{
			name: "SUCCESS: QRIS BRI with required fields",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         "BRI",
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer:           "BRI",
								AcquirerMerchantID: "MID123",
								AcquirerTerminalID: "TID123",
							},
						},
					},
				}, nil)
				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
				qrisSvc.On("CreateManualRegistration", mock.Anything, mock.Anything).Return("reg-id", nil)
				qrisSvc.On("UpdateQrRegistration", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				acquirerMerchantID := "MID123"
				acquirerTerminalID := "TID123"
				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(&qrisModel.Registration{
					Id:                 "reg-id",
					AcquirerMerchantId: &acquirerMerchantID,
					AcquirerTerminalId: &acquirerTerminalID,
					Acquirer:           "BRI",
				}, nil)
				merchantRepo.On("FindMerchantForQrRegistration", mock.Anything, mock.Anything, mock.Anything).Return(&merchantModel.QrisMerchant{}, nil)
				snapCoreRepo.On("QrSyncRegistration", mock.Anything, mock.Anything).Return(nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
						Acquirer:           "BRI",
						AcquirerMerchantID: "MID123",
						AcquirerTerminalID: "TID123",
						MerchantType:       "LARGE",
						CreatedBy:          "test-user",
					},
				},
			},
		},
		{
			name: "ERROR: QRIS BNC (no required fields)",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
							},
						},
					},
				}, nil)
			},
			wantErr: true,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
				},
			},
		},
		{
			name: "SUCCESS: eWallet ShopeePay",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_SHOPEEPAY,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
								ExternalMerchantID: "external-merchant-id",
								ExternalStoreID:    "external-store-id",
							},
						},
					},
				}, nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
						ExternalMerchantID: "external-merchant-id",
						ExternalStoreID:    "external-store-id",
					},
				},
			},
		},
		{
			name: "SUCCESS: eWallet DANA",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_DANA,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
								SubMerchantID: "sub-merchant-id",
							},
						},
					},
				}, nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
						SubMerchantID: "sub-merchant-id",
					},
				},
			},
		},
		{
			name: "SUCCESS: Card",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},

					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{},
						},
					},
				}, nil)
				creditCardRepo.On(
					"GetMIDByAcquirerMID", mock.Anything, "TEST001",
				).Return(&creditcardCoreProcessorModel.MIDResponseData{
					Type: constant.PaymentMethodChannelTypeAggregator,
				}, nil)
				creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(nil, nil)
				creditCardRepo.On("CreateMIDMap", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{
								AcquirerMerchantID: "TEST001",
								ChannelType:        constant.PaymentMethodChannelTypeAggregator,
								SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
									AllowedECICodes: []string{"02", "05"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SUCCESS: Payment method with no PartnerConfig",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "OTHER_PAYMENT_TYPE",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
			},
		},
		{
			name: "SUCCESS: Payment method with empty ChannelConfig",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", ExternalId: "external-id"}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "OTHER_PAYMENT_TYPE",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				ChannelConfig:   &paymentMethodModel.SetupPaymentMethodChannelConfigRequest{},
			},
		},
		{
			name: "SUCCESS: Virtual Account with empty ChannelType fallback",
			modifierMock: func(repo *repositoryMocks.IPaymentMethodRepository, merchantSvc *serviceMocks.IMerchantService, snapCoreRepo *repositoryMocks.ISnapCoreRepository, qrisSvc *serviceMocks.IQrisService, merchantRepo *repositoryMocks.IMerchantRepository, creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				merchantSvc.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{UUID: "merchant-uuid", MID: sql.NullString{String: "MID123", Valid: true}}, nil)
				repo.On(
					"FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Processor: c.SnapCoreProcessor,
						Acquirer:  "BCA",
					},
					ChannelType:       "DEFAULT_CHANNEL",
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				// Only UpdateVirtualAccountConfigPrefix is called since there are no Items to process
				snapCoreRepo.On(
					"UpdateVirtualAccountConfigPrefix",
					mock.Anything,
					mock.Anything,
				).Return(nil)
				repo.On(
					"UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything,
				).Return(nil)
			},
			wantErr: false,
			payload: &paymentMethodModel.SetupPaymentMethodConfigRequest{
				PaymentMethodID: validPaymentMethodID,
				MerchantID:      validMerchantID,
				// Empty ChannelType will fallback to paymentMethod.ChannelType
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test
			repo := repositoryMocks.NewIPaymentMethodRepository(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			merchantSvc := serviceMocks.NewIMerchantService(t)
			creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)
			qrisSvc := serviceMocks.NewIQrisService(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)

			tt.modifierMock(repo, merchantSvc, snapCoreRepo, qrisSvc, merchantRepo, creditCardRepo)

			svc := paymentMethodService.New(
				logger, repo, snapCoreRepo, creditCardRepo,
				paymentMethodService.WithMerchantService(merchantSvc),
				paymentMethodService.WithQrisService(qrisSvc),
				paymentMethodService.WithMerchantRepository(merchantRepo),
				paymentMethodService.WithConfig(&config.Config{}),
			)
			err := svc.SetupConfig(context.Background(), tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigInstallment(t *testing.T) {
	type mocker struct {
		repo               *repositoryMocks.IPaymentMethodRepository
		snapCoreRepo       *repositoryMocks.ISnapCoreRepository
		merchantSvc        *serviceMocks.IMerchantService
		creditCardRepo     *repositoryMocks.ICreditcardCoreProcessorRepository
		qrisSvc            *serviceMocks.IQrisService
		merchantRepo       *repositoryMocks.IMerchantRepository
		installmentPlanSvc *serviceMocks.IInstallmentPlanService

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID: defaultMerchantID,
	}
	// todo:
	// * error get dependent payment method
	// * skipping installment plans that is not belong to merchant

	tests := []struct {
		name        string
		mockSetup   func(*mocker)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "ERROR: Payment method subtype is not card",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype: "INVALID_SUBTYPE",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get active credit card payment method",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype: constant.InstallmentPlanPaymentMethodCard,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("errors"))

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: No active credit card payment method",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype: constant.InstallmentPlanPaymentMethodCard,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Missing installment plan IDs",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype: constant.InstallmentPlanPaymentMethodCard,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{PaymentMethod: paymentModel.PaymentMethod{UUID: "card-pm-id"}},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Installment plan service error",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype:  constant.InstallmentPlanPaymentMethodCard,
						Acquirer: "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{PaymentMethod: paymentModel.PaymentMethod{UUID: "card-pm-id"}},
				}, nil)

				m.installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), c.ErrSomeErrorForUnitTest)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: No installment plans found",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype:  constant.InstallmentPlanPaymentMethodCard,
						Acquirer: "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{PaymentMethod: paymentModel.PaymentMethod{UUID: "card-pm-id"}},
				}, nil)

				m.installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return([]*installmentPlanModel.InstallmentPlan{}, int64(0), nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid installment plan IDs",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype:  constant.InstallmentPlanPaymentMethodCard,
						Acquirer: "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{PaymentMethod: paymentModel.PaymentMethod{UUID: "card-pm-id"}},
				}, nil)

				m.installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return([]*installmentPlanModel.InstallmentPlan{
					{
						UUID:          "valid-plan-id",
						MerchantID:    "",
						Acquirer:      "BCA",
						PaymentMethod: constant.InstallmentPlanPaymentMethodCard,
						Status:        constant.InstallmentPlanStatusActive,
					},
					{
						UUID:          "invalid-plan-id",
						MerchantID:    uuid.Max.String(), // Other merchant plans
						Acquirer:      "BCA",
						PaymentMethod: constant.InstallmentPlanPaymentMethodCard,
						Status:        constant.InstallmentPlanStatusActive,
					},
				}, int64(1), nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"invalid-plan-id"},
						},
					},
				}
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Valid installment setup",
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)

				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						Subtype:  constant.InstallmentPlanPaymentMethodCard,
						Acquirer: "BCA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.repo.On("GetListPaymentMethodMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{PaymentMethod: paymentModel.PaymentMethod{UUID: "card-pm-id"}},
				}, nil)

				m.installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return([]*installmentPlanModel.InstallmentPlan{
					{
						UUID:          "plan-id-1",
						MerchantID:    defaultMerchantID,
						Acquirer:      "BCA",
						PaymentMethod: constant.InstallmentPlanPaymentMethodCard,
						Status:        constant.InstallmentPlanStatusActive,
					},
					{
						UUID:          "plan-id-2",
						MerchantID:    "",
						Acquirer:      "BCA",
						PaymentMethod: constant.InstallmentPlanPaymentMethodCard,
						Status:        constant.InstallmentPlanStatusActive,
					},
				}, int64(2), nil)

				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
							InstallmentPlanIDs: []string{"plan-id-1", "plan-id-2"},
						},
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mocker{
				repo:               repositoryMocks.NewIPaymentMethodRepository(t),
				snapCoreRepo:       repositoryMocks.NewISnapCoreRepository(t),
				merchantSvc:        serviceMocks.NewIMerchantService(t),
				creditCardRepo:     repositoryMocks.NewICreditcardCoreProcessorRepository(t),
				qrisSvc:            serviceMocks.NewIQrisService(t),
				merchantRepo:       repositoryMocks.NewIMerchantRepository(t),
				installmentPlanSvc: serviceMocks.NewIInstallmentPlanService(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo,
				paymentMethodService.WithMerchantService(m.merchantSvc),
				paymentMethodService.WithQrisService(m.qrisSvc),
				paymentMethodService.WithMerchantRepository(m.merchantRepo))

			paymentMethodService.WithInstallmentPlanService(svc, m.installmentPlanSvc)

			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigPaymentQris(t *testing.T) {
	type mocker struct {
		repo           *repositoryMocks.IPaymentMethodRepository
		snapCoreRepo   *repositoryMocks.ISnapCoreRepository
		merchantSvc    *serviceMocks.IMerchantService
		creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository
		qrisSvc        *serviceMocks.IQrisService
		merchantRepo   *repositoryMocks.IMerchantRepository

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID: defaultMerchantID,
		KYCStatus: sql.NullString{
			String: c.KYCStatusApproved,
			Valid:  true,
		},
		ParentID: sql.NullString{
			String: defaultMerchantID,
			Valid:  true,
		},
	}

	prePayload := &paymentMethodModel.SetupPaymentMethodConfigRequest{
		PaymentMethodID: defaultPaymentMethodID,
		MerchantID:      defaultMerchantID,
		ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
	}

	testCases := []struct {
		desc        string
		wantErr     bool
		expectedErr error
		mockSetup   func(m *mocker)
	}{
		{
			desc:    "ERROR: FindMerchantByID not found",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
				m.payload = prePayload
			},
		},
		{
			desc:        "ERROR: FindMerchantByID not found merchant is nill",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrMerchantNotFound),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
				m.payload = prePayload
			},
		},
		{
			desc:        "ERROR: FindMerchantByID not found kyc status not approved",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrMerchantShouldKYC),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{
					UUID: defaultMerchantID,
					KYCStatus: sql.NullString{
						String: c.KYCStatusInReview,
						Valid:  true,
					},
					ParentID: sql.NullString{
						String: uuid.NewString(),
						Valid:  true,
					},
				}, nil)
				m.payload = prePayload
			},
		},
		{
			desc:        "ERROR: FindPaymentMethodByIdAndMerchant error database",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrDatabase, assert.AnError),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
				m.payload = prePayload
			},
		},
		{
			desc:        "ERROR: FindPaymentMethodByIdAndMerchant error data not found",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrPaymentMethodNotFound),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				m.payload = prePayload
			},
		},
		{
			desc:    "SUCCESS: QRIS BRI with all required fields",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
								AcquirerMerchantID: "MID123",
								AcquirerTerminalID: "TID123",
							},
						},
					},
				}, nil)
				m.qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
				m.qrisSvc.On("CreateManualRegistration", mock.Anything, mock.Anything).Return("reg-id", nil)
				m.qrisSvc.On("UpdateQrRegistration", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				acquirerMerchantID := "MID123"
				acquirerTerminalID := "TID123"
				m.qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(&qrisModel.Registration{
					Id:                 "reg-id",
					AcquirerMerchantId: &acquirerMerchantID,
					AcquirerTerminalId: &acquirerTerminalID,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				}, nil)
				m.merchantRepo.On("FindMerchantForQrRegistration", mock.Anything, mock.Anything, mock.Anything).Return(&merchantModel.QrisMerchant{}, nil)
				m.snapCoreRepo.On("QrSyncRegistration", mock.Anything, mock.Anything).Return(nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							AcquirerMerchantID: "MID123",
							AcquirerTerminalID: "TID123",
							MerchantType:       "LARGE",
							CreatedBy:          "test-user",
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: QRIS BNC without required fields",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
							},
						},
					},
				}, nil)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: QRIS BRI missing AcquirerMerchantID validation",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("acquirerMerchantId is required")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							},
						},
					},
				}, nil)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: UpsertPaymentMethodMerchantByIdAndMerchant error for QRIS",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
								AcquirerMerchantID: "MID123",
								AcquirerTerminalID: "TID123",
							},
						},
					},
				}, nil)
				m.qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
				m.qrisSvc.On("CreateManualRegistration", mock.Anything, mock.Anything).Return("reg-id", nil)
				m.qrisSvc.On("UpdateQrRegistration", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				acquirerMerchantID := "MID123"
				acquirerTerminalID := "TID123"
				m.qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer", mock.Anything, mock.Anything, mock.Anything).Return(&qrisModel.Registration{
					Id:                 "reg-id",
					AcquirerMerchantId: &acquirerMerchantID,
					AcquirerTerminalId: &acquirerTerminalID,
					Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
				}, nil)
				m.merchantRepo.On("FindMerchantForQrRegistration", mock.Anything, mock.Anything, mock.Anything).Return(&merchantModel.QrisMerchant{}, nil)
				m.snapCoreRepo.On("QrSyncRegistration", mock.Anything, mock.Anything).Return(nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(assert.AnError)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     paymentConstant.PAYMENT_METHOD_QRIS,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							AcquirerMerchantID: "MID123",
							AcquirerTerminalID: "TID123",
							MerchantType:       "LARGE",
							CreatedBy:          "test-user",
						},
					},
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:           repositoryMocks.NewIPaymentMethodRepository(t),
				snapCoreRepo:   repositoryMocks.NewISnapCoreRepository(t),
				merchantSvc:    serviceMocks.NewIMerchantService(t),
				creditCardRepo: repositoryMocks.NewICreditcardCoreProcessorRepository(t),
				qrisSvc:        serviceMocks.NewIQrisService(t),
				merchantRepo:   repositoryMocks.NewIMerchantRepository(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo, paymentMethodService.WithMerchantService(m.merchantSvc), paymentMethodService.WithQrisService(m.qrisSvc), paymentMethodService.WithMerchantRepository(m.merchantRepo))
			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigCreditCard(t *testing.T) {
	type mocker struct {
		repo           *repositoryMocks.IPaymentMethodRepository
		snapCoreRepo   *repositoryMocks.ISnapCoreRepository
		merchantSvc    *serviceMocks.IMerchantService
		creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository
		qrisSvc        *serviceMocks.IQrisService
		merchantRepo   *repositoryMocks.IMerchantRepository

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID:      defaultMerchantID,
		ShortName: "TEST_MERCHANT",
	}

	testCases := []struct {
		desc        string
		wantErr     bool
		expectedErr error
		mockSetup   func(m *mocker)
	}{
		{
			desc:    "SUCCESS: Credit Card - Create new MID",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND")).Once()
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.MIDResponseData{Uuid: uuid.MustParse(uuid.NewString()), Type: c.CreditCardMidTypeAggregator}, nil).Once()
				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND"))
				m.creditCardRepo.On("CreateMIDMap", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDMapResponseData{}, nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeAggregator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA", "MASTERCARD"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "SUCCESS: Credit Card - Update existing MID with DIRECT type",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				existingMID := &creditcardCoreProcessorModel.MIDResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
					Type: c.CreditCardMidTypeDirect,
				}
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(existingMID, nil)
				existingMidMap := &creditcardCoreProcessorModel.MIDMapResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
				}
				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(existingMidMap, nil)
				m.creditCardRepo.On("UpdateMIDMapPriority", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.UpdateMIDMapResponseData{}, nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeFacilitator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "DIRECT_CARD_MID",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "SUCCESS: Credit Card - Update existing MID",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				existingMID := &creditcardCoreProcessorModel.MIDResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
					Type: c.CreditCardMidTypeAggregator,
				}
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(existingMID, nil)
				existingMidMap := &creditcardCoreProcessorModel.MIDMapResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
				}
				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(existingMidMap, nil)
				m.creditCardRepo.On("UpdateMIDMapPriority", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.UpdateMIDMapResponseData{}, nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeAggregator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "EXISTING_CARD_MID",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "Error: Credit Card - got error when setup different channel type",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				existingMID := &creditcardCoreProcessorModel.MIDResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
					Type: c.CreditCardMidTypeAggregator,
				}
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(existingMID, nil)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeFacilitator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "EXISTING_CARD_MID",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
									ChannelType:        c.PaymentMethodChannelTypeFacilitator,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: Credit Card - Wrong processor",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrProcessorNotRegistered),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: "WRONG_PROCESSOR",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - CreateMID failed",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND"))
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Invalid card-funded payout: MIT config not found",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									CardFundedPayoutType: "CIT",          // NOSONAR
									AcquirerMerchantID:   "CARD_MID_123", // NOSONAR
									PartnerProcessor:     "MPGS",         // NOSONAR
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Invalid card-funded payout: duplicate config for CIT type",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									CardFundedPayoutType: "CIT",          // NOSONAR
									AcquirerMerchantID:   "CARD_MID_123", // NOSONAR
									PartnerProcessor:     "MPGS",         // NOSONAR
								},
								{
									CardFundedPayoutType: "CIT",          // NOSONAR
									AcquirerMerchantID:   "CARD_MID_123", // NOSONAR
									PartnerProcessor:     "MPGS",         // NOSONAR
								},
								{
									CardFundedPayoutType: "MIT",          // NOSONAR
									AcquirerMerchantID:   "CARD_MID_123", // NOSONAR
									PartnerProcessor:     "MPGS",         // NOSONAR
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - Failed to get MID after created",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND"))
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123", // NOSONAR
								},
								{
									AcquirerMerchantID: "CARD_MID_456", // NOSONAR
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - MID still not found after created",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND")).Once()
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, nil).Once()

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - failed to get MID",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - failed to get MID map",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND")).Once()
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.MIDResponseData{Uuid: uuid.MustParse(uuid.NewString()), Type: c.CreditCardMidTypeAggregator}, nil).Once()
				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeAggregator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA", "MASTERCARD"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - failed to Update MID map",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND")).Once()
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.MIDResponseData{Uuid: uuid.MustParse(uuid.NewString()), Type: c.CreditCardMidTypeAggregator}, nil).Once()

				existingMidMap := &creditcardCoreProcessorModel.MIDMapResponseData{
					Uuid: uuid.MustParse(uuid.NewString()),
				}
				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(existingMidMap, nil)
				m.creditCardRepo.On("UpdateMIDMapPriority", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeAggregator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA", "MASTERCARD"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: Credit Card - failed to create MID map",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						Processor: c.CreditCardCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND")).Once()
				m.creditCardRepo.On("CreateMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.CreateMIDResponseData{Created: true, Uuid: uuid.MustParse(uuid.NewString())}, nil)
				m.creditCardRepo.On("GetMIDByAcquirerMID", mock.Anything, mock.Anything).Return(&creditcardCoreProcessorModel.MIDResponseData{Uuid: uuid.MustParse(uuid.NewString()), Type: c.CreditCardMidTypeAggregator}, nil).Once()

				m.creditCardRepo.On("FindMIDMapByMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("NOT_FOUND"))
				m.creditCardRepo.On("CreateMIDMap", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     c.PaymentMethodChannelTypeAggregator,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									AcquirerMerchantID: "CARD_MID_123",
									PartnerProcessor:   "VISA",
									PrincipalAvailable: []string{"VISA", "MASTERCARD"},
									PartnerBaseURL:     "https://api.partner.com",
									IsActive:           true,
									Priority:           1,
								},
							},
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:           repositoryMocks.NewIPaymentMethodRepository(t),
				snapCoreRepo:   repositoryMocks.NewISnapCoreRepository(t),
				merchantSvc:    serviceMocks.NewIMerchantService(t),
				creditCardRepo: repositoryMocks.NewICreditcardCoreProcessorRepository(t),
				qrisSvc:        serviceMocks.NewIQrisService(t),
				merchantRepo:   repositoryMocks.NewIMerchantRepository(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo, paymentMethodService.WithMerchantService(m.merchantSvc), paymentMethodService.WithQrisService(m.qrisSvc), paymentMethodService.WithMerchantRepository(m.merchantRepo), paymentMethodService.WithConfig(&config.Config{}))
			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigVirtualAccount(t *testing.T) {
	type mocker struct {
		repo           *repositoryMocks.IPaymentMethodRepository
		snapCoreRepo   *repositoryMocks.ISnapCoreRepository
		merchantSvc    *serviceMocks.IMerchantService
		creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID: defaultMerchantID,
		MID:  sql.NullString{String: "MERCHANT123", Valid: true},
	}

	testCases := []struct {
		desc        string
		wantErr     bool
		expectedErr error
		mockSetup   func(m *mocker)
	}{
		{
			desc:    "SUCCESS: VA with existing config",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Processor: c.SnapCoreProcessor,
						Acquirer:  "BCA",
					},
					ChannelType:       "API",
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything, mock.Anything).Return([]*snapCoreVaModel.VirtualAccountConfigResponseData{
					{UUID: "existing-config-id"},
				}, nil)
				m.snapCoreRepo.On("UpdateVirtualAccountConfigPrefix", mock.Anything, mock.Anything).Return(nil)
				m.repo.On("UpsertPaymentMethodMerchantByIdAndMerchant", mock.Anything, mock.Anything).Return(nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					ChannelType:     "API",
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
								{
									BINPrefix:  "456",
									Type:       "CLOSED_STATIC",
									StartRange: "1000",
									EndRange:   "9999",
								},
							},
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: VA wrong processor",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrProcessorNotRegistered),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Processor: "WRONG_PROCESSOR",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
				}
			},
		},
		{
			desc:    "ERROR: VA GetVirtualAccountConfig failed",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						Processor: c.SnapCoreProcessor,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				m.snapCoreRepo.On("GetVirtualAccountConfig", mock.Anything, mock.Anything).Return(nil, assert.AnError)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						VirtualAccount: &paymentMethodModel.SetupPaymentMethodPartnerConfigForVARequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForVAObj{
								{
									BINPrefix: "456",
									Type:      "OPEN_STATIC",
								},
							},
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:           repositoryMocks.NewIPaymentMethodRepository(t),
				snapCoreRepo:   repositoryMocks.NewISnapCoreRepository(t),
				merchantSvc:    serviceMocks.NewIMerchantService(t),
				creditCardRepo: repositoryMocks.NewICreditcardCoreProcessorRepository(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo, paymentMethodService.WithMerchantService(m.merchantSvc))
			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigEWallet(t *testing.T) {
	type mocker struct {
		repo           *repositoryMocks.IPaymentMethodRepository
		merchantSvc    *serviceMocks.IMerchantService
		snapCoreRepo   *repositoryMocks.ISnapCoreRepository
		creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID: defaultMerchantID,
	}

	testCases := []struct {
		desc        string
		wantErr     bool
		expectedErr error
		mockSetup   func(m *mocker)
	}{
		{
			desc:        "ERROR: EWallet ShopeePay missing ExternalMerchantID",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("externalMerchantId is required for ShopeePay")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_SHOPEEPAY,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
							ExternalStoreID: "store123",
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: EWallet ShopeePay missing ExternalStoreID",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("externalStoreId is required for ShopeePay")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_SHOPEEPAY,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{
							ExternalMerchantID: "merchant123",
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: EWallet DANA missing SubMerchantID",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("subMerchantId is required for Dana")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						Name: paymentConstant.PAYMENT_METHOD_EWALLET_CHANNEL_DANA,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						EWallet: &paymentMethodModel.SetupPaymentMethodPartnerConfigForEWalletRequest{},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:           repositoryMocks.NewIPaymentMethodRepository(t),
				merchantSvc:    serviceMocks.NewIMerchantService(t),
				snapCoreRepo:   repositoryMocks.NewISnapCoreRepository(t),
				creditCardRepo: repositoryMocks.NewICreditcardCoreProcessorRepository(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo, paymentMethodService.WithMerchantService(m.merchantSvc))
			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetupConfigQRISAdditional(t *testing.T) {
	type mocker struct {
		repo           *repositoryMocks.IPaymentMethodRepository
		snapCoreRepo   *repositoryMocks.ISnapCoreRepository
		merchantSvc    *serviceMocks.IMerchantService
		creditCardRepo *repositoryMocks.ICreditcardCoreProcessorRepository
		qrisSvc        *serviceMocks.IQrisService
		merchantRepo   *repositoryMocks.IMerchantRepository

		payload *paymentMethodModel.SetupPaymentMethodConfigRequest
	}

	defaultMerchantID := uuid.NewString()
	defaultPaymentMethodID := uuid.NewString()

	defaultValidMerchant := &merchantModel.Merchant{
		UUID:       defaultMerchantID,
		ExternalId: "external-merchant-id",
		KYCStatus:  sql.NullString{String: c.KYCStatusApproved, Valid: true},
		ParentID:   sql.NullString{String: defaultMerchantID, Valid: true},
	}

	testCases := []struct {
		desc        string
		wantErr     bool
		expectedErr error
		mockSetup   func(m *mocker)
	}{
		{
			desc:        "ERROR: QRIS BRI missing MerchantType",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("merchantType is required")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							AcquirerMerchantID: "MID123",
							AcquirerTerminalID: "TID123",
							CreatedBy:          "test-user",
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: QRIS BRI missing CreatedBy",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, errors.New("createdBy is required")),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							AcquirerMerchantID: "MID123",
							AcquirerTerminalID: "TID123",
							MerchantType:       "LARGE",
						},
					},
				}
			},
		},
		{
			desc:        "ERROR: QRIS missing store ID for static type",
			wantErr:     true,
			expectedErr: pkgErr.New(responseHttp.HttpErrUnprocessableContent, c.ErrQrisInvalidPartnerConfigRequest),
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:             paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer:         paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
						ActivationMethod: c.PaymentMethodActivationMethodManual,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								AcquirerStoreIDs: []string{}, // Empty to trigger nil pointer protection
							},
						},
					},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer:           paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
							AcquirerMerchantID: "MID123",
							AcquirerTerminalID: "TID123",
							MerchantType:       "LARGE",
							CreatedBy:          "test-user",
							QRType:             c.QrTypeStatic,
							AcquirerStoreIDs:   []string{"store1", "store2"}, // Adding store IDs to trigger validation error
						},
					},
				}
			},
		},
		{
			desc:    "ERROR: QRIS BNC with no validation",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(defaultValidMerchant, nil)
				m.repo.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     paymentConstant.PAYMENT_METHOD_QRIS,
						Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								AcquirerStoreIDs: []string{}, // Initialize to prevent nil pointer
							},
						},
					},
				}, nil)

				m.payload = &paymentMethodModel.SetupPaymentMethodConfigRequest{
					PaymentMethodID: defaultPaymentMethodID,
					MerchantID:      defaultMerchantID,
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
							Acquirer: paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				repo:           repositoryMocks.NewIPaymentMethodRepository(t),
				snapCoreRepo:   repositoryMocks.NewISnapCoreRepository(t),
				merchantSvc:    serviceMocks.NewIMerchantService(t),
				creditCardRepo: repositoryMocks.NewICreditcardCoreProcessorRepository(t),
				qrisSvc:        serviceMocks.NewIQrisService(t),
				merchantRepo:   repositoryMocks.NewIMerchantRepository(t),
			}

			tc.mockSetup(m)

			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			svc := paymentMethodService.New(logger, m.repo, m.snapCoreRepo, m.creditCardRepo, paymentMethodService.WithMerchantService(m.merchantSvc), paymentMethodService.WithQrisService(m.qrisSvc), paymentMethodService.WithMerchantRepository(m.merchantRepo))
			err := svc.SetupConfig(context.Background(), m.payload)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
