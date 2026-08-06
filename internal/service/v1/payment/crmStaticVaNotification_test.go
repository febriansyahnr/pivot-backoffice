package paymentService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentService_CRMStaticVARetryNotification(t *testing.T) {
	ctx := context.Background()
	vaNumber := "88012345678"
	amountValue := "100000"
	amountCurrency := "IDR"

	testCases := []struct {
		name      string
		payload   *paymentModel.CRMStaticVARetryNotificationRequest
		setupMock func(snapCoreRepo *repositoryMocks.ISnapCoreRepository)
		wantErr   bool
	}{
		{
			name: "success - publish static VA notification",
			payload: &paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: vaNumber,
				Amount: commonModel.Amount{
					Value:    amountValue,
					Currency: amountCurrency,
				},
			},
			setupMock: func(snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				snapCoreRepo.On("PublishPayment", mock.Anything, snapPaymentModel.PublishRequest{
					InternalReference: vaNumber,
					PaymentMethod:     constant.UnifiedPaymentMethodVA,
					ForceSuccess:      true,
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    amountValue,
					},
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - va number is required",
			payload: &paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: "",
				Amount: commonModel.Amount{
					Value:    amountValue,
					Currency: amountCurrency,
				},
			},
			setupMock: func(snapCoreRepo *repositoryMocks.ISnapCoreRepository) {},
			wantErr:   true,
		},
		{
			name: "error - amount value is required",
			payload: &paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: vaNumber,
				Amount: commonModel.Amount{
					Value:    "",
					Currency: amountCurrency,
				},
			},
			setupMock: func(snapCoreRepo *repositoryMocks.ISnapCoreRepository) {},
			wantErr:   true,
		},
		{
			name: "error - amount currency is required",
			payload: &paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: vaNumber,
				Amount: commonModel.Amount{
					Value:    amountValue,
					Currency: "",
				},
			},
			setupMock: func(snapCoreRepo *repositoryMocks.ISnapCoreRepository) {},
			wantErr:   true,
		},
		{
			name: "error - snap core publish payment fails",
			payload: &paymentModel.CRMStaticVARetryNotificationRequest{
				VANumber: vaNumber,
				Amount: commonModel.Amount{
					Value:    amountValue,
					Currency: amountCurrency,
				},
			},
			setupMock: func(snapCoreRepo *repositoryMocks.ISnapCoreRepository) {
				publishErr := errors.New("snap core service error")
				snapCoreRepo.On("PublishPayment", mock.Anything, mock.Anything).Return(publishErr)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)

			tc.setupMock(snapCoreRepo)

			service := &PaymentService{
				logger:       logger,
				snapCoreRepo: snapCoreRepo,
			}

			// Execute
			err := service.CRMStaticVARetryNotification(ctx, tc.payload)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			snapCoreRepo.AssertExpectations(t)
		})
	}
}
