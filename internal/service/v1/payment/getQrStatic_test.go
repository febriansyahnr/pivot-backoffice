package paymentService

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetQrMpmStatic(t *testing.T) {
	testCases := []struct {
		name         string
		wantErr      bool
		setupPayload func() *paymentModel.QueryQrMpmStaticRequest
		setupMocks   func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
			mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository)
	}{
		{
			name:    "success query qr mpm static",
			wantErr: false,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
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
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-id").Return(&merchant.Merchant{
					UUID:      "merchant-id",
					Name:      "name",
					ShortName: "",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmStatic", mock.Anything, mock.Anything).Return(&snapCoreModel.QueryQrMpmStaticResponseData{
					DetailData: []snapCoreModel.TransactionHistoryListResponseDetailData{{}, {}, {}},
				}, nil)
			},
		},
		{
			name:    "success query qr mpm static for submerchant",
			wantErr: false,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
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
					"subMerchantId": "sub-merchant-id",
				}
				referenceId := "reference-id"
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "sub-merchant-id").Return(&merchant.Merchant{
					ParentID: sql.NullString{
						String: "merchant-id",
						Valid:  true,
					},
					Name:      "name",
					ShortName: "",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmStatic", mock.Anything, mock.Anything).Return(&snapCoreModel.QueryQrMpmStaticResponseData{
					DetailData: []snapCoreModel.TransactionHistoryListResponseDetailData{{}, {}, {}},
				}, nil)
			},
		},
		{
			name:    "error when get payment",
			wantErr: true,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(nil, fmt.Errorf("error when get payment data by id"))
			},
		},
		{
			name:    "error payment not found",
			wantErr: true,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(nil, nil)
			},
		},
		{
			name:    "error when get merchant data",
			wantErr: true,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
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
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-id").Return(nil, fmt.Errorf("error when find merchant by id"))
			},
		},
		{
			name:    "error when validate payment data",
			wantErr: true,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				metadata := map[string]any{
					"qrType": constant.QrTypeStatic,
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
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-id").Return(&merchant.Merchant{
					Name:      "name",
					ShortName: "",
				}, nil)
			},
		},
		{
			name:    "error when query qr mpm static to snap core",
			wantErr: true,
			setupPayload: func() *paymentModel.QueryQrMpmStaticRequest {
				return &paymentModel.QueryQrMpmStaticRequest{
					ReferenceId: "reference-id",
				}
			},
			setupMocks: func(mockPaymentRepo *repositoryMocks.IPaymentRepository,
				mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
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
				mockPaymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, mock.Anything, "reference-id").Return(&paymentModel.Payment{
					ReferenceID: &referenceId,
					Metadata:    &metadata,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constantPayment.PAYMENT_METHOD_QRIS,
					},
				}, nil)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-id").Return(&merchant.Merchant{
					UUID:      "merchant-id",
					Name:      "",
					ShortName: "short-name",
				}, nil)
				mockSnapCoreRepo.On("QueryQrMpmStatic", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error when query qr mpm static to snapcore"))
			},
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

			svc := New(mockPaymentRepo, mockLogger, mockSnapCore, mockCustomerRepo, mockMerchantRepo, mockPaymentMethodRepo, nil)
			_, err := svc.GetQrMpmStatic(ctx, tc.setupPayload(), "merchant-id")

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
