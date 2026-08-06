package notificationService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/notification"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendFailedWithdrawalAlert(t *testing.T) {
	logger := mocks.NewILogger(t)
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	
	config := &config.Config{
		SlackConfig: config.SlackConfig{
			WithdrawalAlertWebHookURL: "https://hooks.slack.com/services/webhook-url",
		},
	}

	service := New(config, logger, rmq)
	ctx := context.Background()

	tests := []struct {
		name      string
		request   *withdrawal.FailedWithdrawalAlertRequest
		setupMock func()
		wantErr   error
	}{
		{
			name: "SUCCESS: Send alert with all fields",
			request: &withdrawal.FailedWithdrawalAlertRequest{
				WithdrawalID:              "withdrawal-123",
				MerchantID:                "merchant-456",
				BalanceName:               "Primary Balance",
				BeneficiaryAccountName:    "John Doe",
				BeneficiaryAccountNo:      "1234567890",
				BeneficiaryAccountBankName: "Bank Central Asia",
				Amount:                    100000.50,
				WithdrawType:              "manual",
				Status:                    "failed",
				Reason:                    "Insufficient funds",
				AlertTitle:                "Withdrawal Failed",
			},
			setupMock: func() {
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Send alert with minimal fields",
			request: &withdrawal.FailedWithdrawalAlertRequest{
				WithdrawalID: "withdrawal-456",
				MerchantID:   "merchant-789",
				Amount:       50000,
				Status:       "failed",
				AlertTitle:   "Alert",
			},
			setupMock: func() {
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "ERROR: RabbitMQ publish fails",
			request: &withdrawal.FailedWithdrawalAlertRequest{
				WithdrawalID: "withdrawal-789",
				MerchantID:   "merchant-999",
				Amount:       75000,
				Status:       "failed",
				AlertTitle:   "Failed Alert",
			},
			setupMock: func() {
				publishErr := errors.New("rabbitmq connection failed")
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(publishErr).Once()
				logger.On("Error", mock.Anything, "error when publish to slack queue", mock.Anything).Return().Once()
			},
			wantErr: errors.New("rabbitmq connection failed"),
		},
		{
			name: "SUCCESS: Send alert with zero amount",
			request: &withdrawal.FailedWithdrawalAlertRequest{
				WithdrawalID: "withdrawal-zero",
				MerchantID:   "merchant-zero",
				Amount:       0,
				Status:       "failed",
				AlertTitle:   "Zero Amount Alert",
			},
			setupMock: func() {
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			err := service.SendFailedWithdrawalAlert(ctx, tt.request)
			
			assert.Equal(t, tt.wantErr, err)
			rmq.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}