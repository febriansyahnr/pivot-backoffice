package notificationService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendVccSettlementTransactionAlert(t *testing.T) {
	logger := mocks.NewILogger(t)
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)

	config := &config.Config{
		SlackConfig: config.SlackConfig{
			VccSettlementTransactionInquiryAlertWebhookURL: "https://hooks.slack.com/services/webhook-url",
		},
		VccSlackRecipient: config.VccSlackRecipientConfig{
			DefaultRecipient: []string{"recipient-1", "recipient-2"},
		},
	}

	service := New(config, logger, rmq)
	ctx := context.Background()

	tests := []struct {
		name      string
		request   *vccSettlement.VccTransactionInquiryAlert
		setupMock func()
		wantErr   error
	}{
		{
			name: "SUCCESS: Send alert with all fields",
			request: &vccSettlement.VccTransactionInquiryAlert{
				Title:       "title",
				Recipient:   config.VccSlackRecipient.DefaultRecipient,
				Description: "description",
				RcnId:       "rcn-id",
				PostingDate: "20200101",
			},
			setupMock: func() {
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "ERROR: RabbitMQ publish fails",
			request: &vccSettlement.VccTransactionInquiryAlert{
				Title:       "title",
				Recipient:   config.VccSlackRecipient.DefaultRecipient,
				Description: "description",
				RcnId:       "rcn-id",
				PostingDate: "20200101",
			},
			setupMock: func() {
				publishErr := errors.New("rabbitmq connection failed")
				rmq.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(publishErr).Once()
				logger.On("Error", mock.Anything, "error when publish to slack queue", mock.Anything).Return().Once()
			},
			wantErr: errors.New("rabbitmq connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := service.SendVccSettlementTransactionAlert(ctx, tt.request)

			assert.Equal(t, tt.wantErr, err)
			rmq.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
