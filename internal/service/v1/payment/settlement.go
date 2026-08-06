package paymentService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) tryProcessSettlementCutOff(ctx context.Context, transactionID, feeID, merchantID string, date time.Time, config *merchantModel.SettlementConfig) bool {

	if config == nil || config.CutOff == nil {
		return false
	}

	// Time processing using WIB time zone
	date = date.In(loc)

	if cutoff, err := config.CutOff.Window.IsCutOffTime(date); err != nil {
		s.logger.Error(ctx, "Failed during cut-off time check", logger.Error(err))
		return false

	} else if !cutoff {
		return false
	}

	settlementTime, _ := config.GetSettlementTime(date) // err is ignored because already handled in previous lines

	payload := settlementModel.ProcessSettlementRequest{
		TransactionID:    transactionID,
		FeeTransactionID: feeID,
		MerchantID:       merchantID,
		Type:             constant.SettlementTransaction,
	}
	duration := settlementTime.Sub(date)

	s.logger.Info(
		ctx, fmt.Sprintf("Transaction ID %s is within settlement cut-off and will be settled at %s", transactionID, settlementTime),
		logger.String("duration", duration.String()), logger.Any("payload", payload),
	)
	if err := s.rabbitMqExt.PublishWithDelay(ctx, rabbitMqExt.SettlementProcessingRoutingKey, payload, duration); err != nil {
		s.logger.Error(ctx, "Failed while publishing additional payment settlement delay message", logger.Error(err))
		return false
	}

	ids := []string{transactionID, feeID}
	updateRequest := orchestratorModel.UpdateSettlementDetailRequest{
		EstimateSettlementAt: util.ValueToPtr(settlementTime.In(time.UTC)),
	}
	if err := s.accountTransactionRepo.UpdateSettlementDetailByIDs(ctx, ids, updateRequest); err != nil {
		s.logger.Warn(ctx, "Settlement detail update failed, but the process is considered successful as the message was published", logger.Error(err))
	}
	return true
}
