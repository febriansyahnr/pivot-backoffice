package paymentService

import (
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTryProcessSettlementCutOff(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := &PaymentService{
		rabbitMqExt: rmq, logger: logger, accountTransactionRepo: accountTransactionRepo,
	}

	var (
		transactionID = "e0d2624a-6f01-41c4-9825-900489a09a6a"
		feeID         = "dce7bd59-a81e-450e-95f8-6bba4b855a40"
		merchantID    = "5d210229-5652-4ad2-b3b3-8e7af22be890"
		ids           = []string{transactionID, feeID}
	)

	settlementConfig := &merchant.SettlementConfig{
		Type: "T+1",
		CutOff: &merchant.SettlementConfigCutOff{
			Window: merchant.SettlementConfigCutOffWindow{
				StartTime: "22:00:00", // NOSONAR
				EndTime:   "06:00:00", // NOSONAR
			},
			Deferral: merchant.SettlementConfigCutOffDeferral{
				OffsetDays:    1,          // NOSONAR
				ExecutionTime: "07:00:00", // NOSONAR
			},
		},
	}

	tests := []struct {
		name       string
		date       time.Time
		config     *merchant.SettlementConfig
		setupMock  func()
		wantResult bool
	}{
		{
			name:       "SUCCESS:No cut-off config (nil config)",
			wantResult: false,
		},
		{
			name:       "SUCCESS:No cut-off config (nil cut-off)",
			config:     &merchant.SettlementConfig{},
			wantResult: false,
		},
		{
			name: "ERROR:Invalid window config",
			date: time.Date(2026, 2, 10, 14, 21, 19, 0, time.UTC),
			config: &merchant.SettlementConfig{
				CutOff: &merchant.SettlementConfigCutOff{},
			},
			setupMock: func() {
				logger.On(
					"Error", mock.Anything, "Failed during cut-off time check", mock.Anything,
				).Once().Return()
			},
			wantResult: false,
		},
		{
			name:       "SUCCESS:Outside the cut-off time",
			date:       time.Date(2026, 2, 10, 14, 21, 19, 0, time.UTC),
			config:     settlementConfig,
			wantResult: false,
		},
		{
			name:   "ERROR:Publish settlement process",
			date:   time.Date(2026, 2, 10, 15, 56, 19, 0, time.UTC),
			config: settlementConfig,
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, fmt.Sprintf("Transaction ID %s is within settlement cut-off and will be settled at 2026-02-12 07:00:00 +0700 WIB", transactionID), mock.Anything, mock.Anything,
				).Once().Return()
				rmq.On(
					"PublishWithDelay", mock.Anything, rabbitMqExt.SettlementProcessingRoutingKey, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
				logger.On(
					"Error", mock.Anything, "Failed while publishing additional payment settlement delay message", mock.Anything,
				).Once().Return()
			},
			wantResult: false,
		},
		{
			name:   "ERROR:Update settlement details",
			date:   time.Date(2026, 2, 10, 15, 56, 19, 0, time.UTC),
			config: settlementConfig,
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, fmt.Sprintf("Transaction ID %s is within settlement cut-off and will be settled at 2026-02-12 07:00:00 +0700 WIB", transactionID), mock.Anything, mock.Anything,
				).Once().Return()
				rmq.On(
					"PublishWithDelay", mock.Anything, rabbitMqExt.SettlementProcessingRoutingKey, mock.Anything, mock.Anything,
				).Once().Return(nil)
				accountTransactionRepo.On(
					"UpdateSettlementDetailByIDs", mock.Anything, ids, mock.Anything,
				).Once().Return(assert.AnError)
				logger.On(
					"Warn", mock.Anything, "Settlement detail update failed, but the process is considered successful as the message was published", mock.Anything,
				).Once().Return()
			},
			wantResult: true,
		},
		{
			name:   "SUCCESS:Before midnight",
			date:   time.Date(2026, 2, 10, 16, 59, 59, 0, time.UTC),
			config: settlementConfig,
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, fmt.Sprintf("Transaction ID %s is within settlement cut-off and will be settled at 2026-02-12 07:00:00 +0700 WIB", transactionID), mock.Anything, mock.Anything,
				).Once().Return()
				rmq.On("PublishWithDelay", mock.Anything, rabbitMqExt.SettlementProcessingRoutingKey, mock.Anything, mock.Anything).Return(nil)
				accountTransactionRepo.On("UpdateSettlementDetailByIDs", mock.Anything, ids, mock.Anything).Return(nil)
			},
			wantResult: true,
		},
		{
			name:   "SUCCESS:After midnight",
			date:   time.Date(2026, 2, 10, 20, 56, 19, 0, time.UTC),
			config: settlementConfig,
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, fmt.Sprintf("Transaction ID %s is within settlement cut-off and will be settled at 2026-02-13 07:00:00 +0700 WIB", transactionID), mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantResult, service.tryProcessSettlementCutOff(
				t.Context(), transactionID, feeID, merchantID, test.date, test.config,
			))
			rmq.AssertExpectations(t)
			logger.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
		})
	}
}
