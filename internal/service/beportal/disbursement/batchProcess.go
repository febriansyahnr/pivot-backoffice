package disbursementService

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pbCommon "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

func (s *DisbursementService) BatchProcessDisbursement(ctx context.Context, request *disbursementModel.BatchProcessDisbursementRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/BatchProcessDisbursement")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(ctx, "Panic recovery from BatchProcessDisbursement", logger.Error(fmt.Errorf("%v", r)))
			// TODO: send to slack
		}
	}()

	now := time.Now()
	defer monitor.WriteAndSend(
		ctx, "disbursement-batch-process", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("bulk_id:%s", request.BulkID),
				fmt.Sprintf("total:%d", len(request.DisbursementIDs)),
				"proc_identifier:batch-process",
			}
		},
	)

	if s.batchProcessWP == nil {
		s.newBatchProcessWP()
	}

	var (
		wg sync.WaitGroup
	)

	// TODO: Need to highly monitor this, because there will be so much span in go routine
	for _, disbursementID := range request.DisbursementIDs {
		wg.Add(1)

		// Invoke create disbursement
		_ = s.batchProcessWP.Invoke(batchProcessWPData{
			ctx:            ctx,
			wg:             &wg,
			disbursementID: disbursementID,
		})
	}

	wg.Wait()

	// Skip to update parent status for multiple approval for single transaction
	if request.BulkID == "" {
		return nil
	}

	// Trigger check count in progress, if count = 0 then update bulkDisbursement status to DONE
	if errUpdate := s.updateParentStatusAndSendCallback(ctx, request.BulkID, request.DisbursementIDs[0]); errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (s *DisbursementService) sendCallback(ctx context.Context, bulkID, merchantID, bulkStatus, event string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/sendCallback")
	defer segment.End()

	disbursementDataRequest := &pb.DisbursementCallbackRequest{
		MerchantId: merchantID,
	}

	// Count Disbursement Total
	countDisbursement := s.disbursementRepo.CountByBulkID(ctx, bulkID)

	// Read feature flag by merchant for single payout callback
	if countDisbursement == 1 && constant.IsSinglePayoutCallbackWhitelistedForMerchant(merchantID) {
		disbursements, err := s.disbursementRepo.GetAllDisbursementByBulkID(ctx, bulkID)
		if err != nil {
			s.logger.Error(ctx, "error getting disbursements by bulk id", logger.Error(err))
			return err
		} else if len(disbursements) != 1 {
			s.logger.Warn(ctx, "disbursements is not single payout", logger.Error(err))
			return constant.ErrPayoutIsNotSingle
		} else if disbursements[0] == nil {
			s.logger.Warn(ctx, "disbursements is nil", logger.Error(err))
			return constant.ErrDisbursementNotFound
		}

		disbursement := disbursements[0]
		disbursementDataRequest = &pb.DisbursementCallbackRequest{
			Uuid:       disbursement.UUID, // Use disbursement UUID for single payout
			MerchantId: merchantID,
			Payout: &pb.PayoutCallbackSingleObject{
				ReferenceId: disbursement.ReferenceID,
				Amount: &pbCommon.AmountDouble{
					Currency: disbursement.Currency,
					Value:    disbursement.TotalAmount.InexactFloat64(),
				},
			},
		}

		disbursementDataRequest.Payout.Status = BuildChildStatus(*disbursement)
		if (disbursementDataRequest.Payout.Status == constant.StatusFailed) || (disbursementDataRequest.Payout.Status == constant.DisbursementReasonTypeDelayed) {
			disbursementDataRequest.Payout.Reason = BuildReason(*disbursement)
		}

		// Build event from payout status
		switch disbursementDataRequest.Payout.Status {
		case constant.StatusSuccess:
			event = constant.CallbackEventPayoutSuccess
		case constant.StatusFailed:
			event = constant.CallbackEventPayoutFailed
		case constant.DisbursementReasonTypeDelayed:
			event = constant.CallbackEventPayoutDelayed
		case constant.DisbursementReasonTypeCancelled:
			event = constant.CallbackEventPayoutCancelled
		}

	} else {
		payoutResult := s.getDisbursementSummaryByBulkID(ctx, bulkID)
		disbursementDataRequest = &pb.DisbursementCallbackRequest{
			Uuid:       bulkID,
			MerchantId: merchantID,
			PayoutResults: &pb.DisbursementObject{
				TotalPendingCount:    float64(payoutResult.TotalPendingCount),
				TotalPendingAmount:   payoutResult.TotalPendingAmount,
				TotalSuccessCount:    float64(payoutResult.TotalSuccessCount),
				TotalSuccessAmount:   payoutResult.TotalSuccessAmount,
				TotalFailedCount:     float64(payoutResult.TotalFailedCount),
				TotalFailedAmount:    payoutResult.TotalFailedAmount,
				TotalCancelledCount:  float64(payoutResult.TotalCancelledCount),
				TotalCancelledAmount: payoutResult.TotalCancelledAmount,
			},
			Status: bulkStatus,
		}
	}

	// Callback Body Content

	anyWrapper, err := anypb.New(disbursementDataRequest)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(constant.DisbursementCallbackEventLockFmt, merchantID, bulkID, event)
	s.logger.Info(ctx, "acquiring disbursement callback lock key", logger.String("key", key))
	isLockAcquired, err := s.redisExt.SetNX(ctx, key, "1", constant.CallbackEventLockTTL).Result()
	if err != nil {
		s.logger.Error(ctx, "error acquire disbursement callback lock event", logger.Error(err), logger.String("key", key))
		return err
	}
	if !isLockAcquired {
		s.logger.Info(ctx, "failed to acquire disbursement callback lock event", logger.String("key", key))
		return nil
	}

	// Callback Delivery Data Format
	recipientMerchantIds, _ := s.disbursementRepo.GetMerchantIDsForPayoutCallback(ctx, bulkID)
	if len(recipientMerchantIds) == 0 {
		recipientMerchantIds = append(recipientMerchantIds, merchantID)
	}
	for _, recipientId := range recipientMerchantIds {
		callbackRequest := &pb.ProcessCallbackRequest{
			Name:       constant.CallbackNameDisbursement,
			Event:      event,
			MerchantId: recipientId,
			Request:    anyWrapper,
		}

		_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)
	}
	return nil
}

func (s *DisbursementService) updateParentStatusAndSendCallback(ctx context.Context, bulkID string, disbursementID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/updateParentStatusAndSendCallback")
	defer segment.End()

	if bulkID == "" {
		return nil
	}

	// Trigger check count in progress, if count = 0 then update bulkDisbursement status to DONE
	countByBulkID := s.disbursementRepo.CountStatusInProgressByBulkID(ctx, bulkID)
	if countByBulkID == 0 {
		if err := s.disbursementRepo.UpdateBulkDisbursementStatusByID(ctx, bulkID, constant.BulkDisbursementStatusDone); err != nil {
			return err
		}

		sampleDisbursement, err := s.disbursementRepo.FindByID(ctx, disbursementID)
		if err != nil {
			return err
		}

		if sampleDisbursement == nil {
			return constant.ErrDisbursementNotFound
		}

		// Send callback only when created from OPEN API
		if sampleDisbursement.CreatedFrom != nil && *sampleDisbursement.CreatedFrom == constant.DisbursementCreatedFromOpenApi {
			if err = s.sendCallback(ctx, bulkID, sampleDisbursement.MerchantID, constant.BulkDisbursementStatusDone, constant.CallbackEventPayoutDone); err != nil {
				return err
			}
		}
	}

	return nil
}
