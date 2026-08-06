package paymentService

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetQrMpmDynamic(t *testing.T) {
	testCases := []struct {
		name        string
		uuid        string
		referenceId string
		setupMocks  func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
			mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository,
			mockMerchantRepo *repositoryMocks.IMerchantRepository)
		wantErr bool
	}{
		{
			name: "success query qr mpm dynamic by uuid with status success",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository,
				mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "merchant-id",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmDynamic", mock.Anything, "snap-core-uuid").Return(&snapCoreModel.QueryQrMpmDynamicResponseData{
					Status: "SUCCESS",
					Amount: commonModel.Amount{
						Value:    "10000",
						Currency: "IDR",
					},
				}, nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success query qr mpm dynamic by uuid with status success as submerchant",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
					"subMerchantId": "sub-merchant-id",
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "sub-merchant-id").Return(&merchant.Merchant{
					ParentID: sql.NullString{
						String: "merchant-id",
						Valid:  true,
					},
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmDynamic", mock.Anything, "snap-core-uuid").Return(&snapCoreModel.QueryQrMpmDynamicResponseData{
					Status: "SUCCESS",
					Amount: commonModel.Amount{
						Value:    "10000",
						Currency: "IDR",
					},
				}, nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error on database when query qr mpm dynamic by uuid with status success as submerchant ",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
					"subMerchantId": "sub-merchant-id",
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "sub-merchant-id").Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error merchant not found when query qr mpm dynamic by uuid with status success as submerchant ",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
					"subMerchantId": "sub-merchant-id",
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "sub-merchant-id").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:        "error query qr mpm dynamic by referenceId",
			referenceId: "reference-id",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(nil, constant.ErrPaymentNotFound)
			},
			wantErr: true,
		},
		{
			name:        "error data not found query qr mpm dynamic by referenceId",
			referenceId: "reference-id",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error validate payment data",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeStatic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "STATIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "success query qr mpm dynamic by uuid and status is already success",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "merchant-id",
				}, nil)
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_SUCCESS,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "error query qr mpm dynamic to snap core",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "merchant-id",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmDynamic", mock.Anything, "snap-core-uuid").Return(nil, errors.New("error query qr mpm dynamic"))
			},
			wantErr: true,
		},
		{
			name: "success query qr mpm dynamic by uuid and query status is still pending",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "merchant-id",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmDynamic", mock.Anything, "snap-core-uuid").Return(&snapCoreModel.QueryQrMpmDynamicResponseData{
					Status: "PENDING",
					Amount: commonModel.Amount{
						Value:    "10000",
						Currency: "IDR",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "error update query qr mpm dynamic by uuid with status void",
			uuid: "uuid",
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeDynamic,
					"snapCore": map[string]any{
						"uuid":   "snap-core-uuid",
						"status": "PENDING",
						"qrType": "DYNAMIC",
						"amount": map[string]any{
							"value":    "10000",
							"currency": "IDR",
						},
					},
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "uuid").Return(&paymentModel.Payment{
					UUID:        "uuid",
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
					Status: constantPayment.PAYMENT_STATUS_PENDING,
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{
					UUID: "merchant-id",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmDynamic", mock.Anything, "snap-core-uuid").Return(&snapCoreModel.QueryQrMpmDynamicResponseData{
					Status: "CANCELLED",
					Amount: commonModel.Amount{
						Value:    "10000",
						Currency: "IDR",
					},
				}, nil)
				mockPaymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Return(errors.New("error update payment data"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockCustomerRepo := repositoryMocks.NewICustomerRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockSnapCore := repositoryMocks.NewISnapCoreRepository(t)
			mockMerchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockPaymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)

			tc.setupMocks(mockPaymentRepo, mockSnapCore, mockMerchantRepo)

			merchantInfo := &merchant.MerchantAuthTokenClaims{
				MerchantId: "merchant-id",
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, constant.CtxMerchantInfo, merchantInfo)

			svc := New(mockPaymentRepo, mockLogger, mockSnapCore, mockCustomerRepo, mockMerchantRepo, mockPaymentMethodRepo, nil,
				WithConfig(&config.Config{
					MerchantPortalConfig: config.MerchantPortalConfig{
						PaymentSimulationPatternURL: "https://dashboard-stg.harsya.com/simulation/payment/%s",
					},
				}))
			_, err := svc.GetQrMpmDynamic(ctx, tc.uuid, tc.referenceId, merchantInfo.MerchantId)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockCustomerRepo.AssertExpectations(t)
			mockSnapCore.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
		})
	}
}
