package disbursementService

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-redsync/redsync/v4"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
)

func (s *DisbursementService) Approve(ctx context.Context, request *disbursementModel.ApproveRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/Approve")
	defer segment.End()

	if len(request.ApproveAction) == 0 {
		return nil
	}

	var disbursementIDs []string
	for _, approveAction := range request.ApproveAction {
		disbursementIDs = append(disbursementIDs, approveAction.DisbursementID)
	}

	// Set Distributed Lock
	mutex := s.redisExt.NewMutex(
		"backend-portal:merchant-balances:"+request.MerchantID+":deduct:disbursement",
		redsync.WithExpiry(60*time.Second),
		redsync.WithRetryDelay(80*time.Millisecond),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
	if err := mutex.LockContext(ctx); err != nil {
		return pkgErrs.New(response.HttpErrRequest, err)
	}
	unlockProcess := func() {
		if _, err := mutex.UnlockContext(context.WithoutCancel(ctx)); err != nil {
			s.logger.Warn(ctx, "Failed unlock distributed lock", logger.Error(err))
		}
	}

	trxCtx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		unlockProcess()
		return err
	}

	request.IsCompleted = false
	defer func() {
		if request.IsCompleted {
			return
		}
		if e := s.disbursementRepo.RollbackTransaction(trxCtx); e != nil {
			s.logger.Warn(ctx, "Rollback approval transaction failed", logger.Error(e))
		}
		unlockProcess()
	}()

	disbursementInfo := map[string]any{
		"merchantId":      request.MerchantID,
		"from":            request.CreatedFrom,
		"totalAmount":     request.TotalAmount, // Without Fee
		"disbursementIds": disbursementIDs,
	}

	// Update disbursement status from WAITING to APPROVED
	if err = s.disbursementRepo.ApproveInBulk(trxCtx, disbursementIDs, request.ApprovedBy); err != nil {
		if errors.Is(err, constant.ErrNoRowsAffected) {
			s.logger.Info(
				ctx, "Disbursement transactions have gone through the approval process", logger.Any("details", disbursementInfo),
			)
			return pkgErrs.New(response.HttpErrRequest, constant.ErrDisbursementStatusAlreadyApproved)
		}
		return err
	}

	// Record status history for approved disbursements using workerpool pattern
	if s.statusHistoryWP == nil {
		s.newStatusHistoryWP()
	}
	s.recordStatusHistoryApproved(ctx, request.ApprovedBy, disbursementIDs)

	if valid, err := s.validateBalanceAndUpdateIfNotValid(ctx, trxCtx, disbursementIDs, request); err != nil {
		return err

	} else if !valid {
		// Commit with Insufficient Balance disbursement status
		if err = s.disbursementRepo.CommitTransaction(trxCtx); err != nil {
			return err
		}
		unlockProcess()
		request.IsCompleted = true
		_ = s.self.DecrDailyTransactionLimit(ctx, request.MerchantID, request.TotalAmount)

		s.logger.Info(ctx, "Merchant balance is insufficient", logger.Any("details", disbursementInfo))
		return pkgErrs.New(response.HttpErrRequest, constant.ErrInsufficientBalance)
	}

	// Payout Cut-off Time Window
	merchantId := request.MerchantID
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		merchantId = parentMerchantId
	}
	payoutCutOffTime, err := s.self.GetCutOffTimeStatus(trxCtx, time.Now().UTC(), merchantId, nil)
	if err != nil {
		s.logger.Error(ctx, "Get payout cut-off time window", logger.Error(err))
		return err
	}

	// Update bulk transactions status to IN_PROGRESS
	if request.BulkID != "" {
		if err = s.disbursementRepo.UpdateBulkDisbursementStatusByID(trxCtx, request.BulkID, constant.BulkDisbursementStatusInProgress); err != nil {
			return err
		}
	}

	if err = s.disbursementRepo.CommitTransaction(trxCtx); err != nil {
		return err
	}
	parentCtx := ctx

	// Removing context cancellation to allow the process to run independently serves as a temporary workaround until a more efficient process is implemented in the future.
	// As a result, code execution after this line will continue to run even if the HTTP request is disconnected or times out by the client.
	ctx = context.WithoutCancel(ctx)

	// The transaction initiation process begins by recording the transaction with a 'pending' status, followed by beneficiary validation.
	approvalValidations := s.processBatchApproval(ctx, disbursementIDs, payoutCutOffTime)

	unlockProcess()
	request.IsCompleted = true

	// Trigger process disbursement in async (bulk or multiple approval)
	s.triggerPublishBatchProcess(ctx, request.BulkID, disbursementIDs, payoutCutOffTime)

	if err := parentCtx.Err(); err != nil {
		message := "Parent context is canceled (timeout / request closed); proceeding with transaction for data consistency."
		segment.SetAttributes(attribute.String("bulk.id", request.BulkID), attribute.String("message", message))
		s.logger.Warn(ctx, message, logger.String("bulkId", request.BulkID), logger.Any("approvalValidations", approvalValidations), logger.Error(err))
	}

	// When all disbursements are approved are got error beneficiary payout limit then should return error beneficiary payout limit
	if len(approvalValidations) == len(disbursementIDs) {
		_ = s.self.DecrDailyTransactionLimit(ctx, request.MerchantID, request.TotalAmount)
		return pkgErrs.New(response.HttpErrDailyLimitReached, constant.ErrBeneficiaryLimitRestrictions)
	}

	approvalResErr := disbursementModel.ApprovalResultErr{}
	for _, validation := range approvalValidations {
		if validation.Error == nil {
			continue
		}

		approvalResErr.BeneficiaryLimitExceeded = append(approvalResErr.BeneficiaryLimitExceeded, validation)
		_ = s.self.DecrDailyTransactionLimit(ctx, request.MerchantID, validation.Amount)
	}

	if len(approvalResErr.BeneficiaryLimitExceeded) > 0 {
		return &approvalResErr
	}

	return nil
}

func (s *DisbursementService) processBatchApproval(ctx context.Context, disbursementIDs []string, payoutCutOffTime *disbursementModel.CutOffTimeStatusResponse) []disbursementModel.ApprovalValidation {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/processBatchApproval")
	defer segment.End()

	segment.SetAttributes(attribute.Int("payout.count", len(disbursementIDs)), attribute.String("trace.id", logger.GetTraceID(ctx)))

	approvalValidations := []disbursementModel.ApprovalValidation{} // currently used for beneficiary payout limit
	cutOffTimeStatusOngoing := (payoutCutOffTime.Status == constant.DisbursementCutOffTimeStatusOngoing)

	wg, mx := new(sync.WaitGroup), new(sync.Mutex)
	for _, id := range disbursementIDs {
		wg.Add(1)
		_ = s.batchApprovalWP.Invoke(
			&batchApprovalWPData{ctx, wg, id, cutOffTimeStatusOngoing, payoutCutOffTime.ProcessedAt, mx, &approvalValidations},
		)
	}
	wg.Wait()

	return approvalValidations
}

func (s *DisbursementService) recordStatusHistoryApproved(ctx context.Context, approvedBy string, disbursementIDs []string) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/recordStatusHistoryApproved")
	defer segment.End()

	wg := &sync.WaitGroup{}
	segment.SetAttributes(attribute.Int("payout.count", len(disbursementIDs)), attribute.String("trace.id", logger.GetTraceID(ctx)))

	for _, disbursementID := range disbursementIDs {
		wg.Add(1)

		_ = s.statusHistoryWP.Invoke(&statusHistoryWPData{
			ctx:            ctx,
			wg:             wg,
			disbursementId: disbursementID,
			actor:          approvedBy,
			statusType:     constant.DisbursementStatusHistoryApproved,
		})
	}

	wg.Wait()
}

func (s *DisbursementService) triggerPublishBatchProcess(ctx context.Context, bulkID string, disbursementIDs []string, payoutCutOffTime *disbursementModel.CutOffTimeStatusResponse) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/triggerPublishBatchProcess")
	defer segment.End()

	wg := sync.WaitGroup{}
	chunkSize := constant.BulkDisbursementMaxDataRequestPerBatch
	segment.SetAttributes(attribute.Int("payout.count", len(disbursementIDs)), attribute.String("trace.id", logger.GetTraceID(ctx)))

	routingKey := rabbitMqExt.BulkDisbursementBatchProcessRoutingKey
	if payoutCutOffTime != nil && payoutCutOffTime.Status == constant.DisbursementCutOffTimeStatusOngoing {
		routingKey = rabbitMqExt.BulkDisbursementBatchDelayTransferRoutingKey
		ctx = context.WithValue(ctx, constant.CtxRabbitMQExpiration, payoutCutOffTime.ProcessedAt.Sub(time.Now().UTC()).Milliseconds())
	}

	// Looping batch, then send to rmq
	for i := 0; i < len(disbursementIDs); i += chunkSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()

			end := start + chunkSize
			if end > len(disbursementIDs) {
				end = len(disbursementIDs)
			}

			batchRequest := &pb.BatchProcessDisbursementRequest{
				BulkId:          bulkID,
				DisbursementIds: disbursementIDs[start:end],
			}
			payload, _ := proto.Marshal(batchRequest)

			_ = s.rabbitMqExt.Publish(ctx, routingKey, nil, payload)
		}(i)
	}

	wg.Wait()
}

func (s *DisbursementService) validateBalanceAndUpdateIfNotValid(ctx, trxCtx context.Context, disbursementIDs []string, request *disbursementModel.ApproveRequest) (bool, error) {
	trxCtx, segment := otelTracer.Start(trxCtx, "internal/service/v1/disbursement/validateBalanceAndUpdateIfNotValid")
	defer segment.End()

	// Validate balance first
	valid := s.ValidateBalance(trxCtx, &disbursementModel.ValidateBalanceRequest{
		DisbursementIDs: disbursementIDs,
		MerchantID:      request.MerchantID,
	})

	// If mock insufficient balance, set valid to false
	mockInsufficientBalance := trxCtx.Value(constant.CtxMockInsufficientBalanceMerchant)
	if mockInsufficientBalanceValue, ok := mockInsufficientBalance.(bool); ok && mockInsufficientBalanceValue {
		valid = false
	}

	if !valid {
		// Update reason type = INSUFFICIENT_BALANCE
		err := s.disbursementRepo.UpdateReasonByIDs(
			trxCtx, disbursementIDs, constant.DisbursementReasonTypeInsufficientBalance, util.ToTitle(constant.DisbursementReasonTypeInsufficientBalance),
		)
		if err != nil {
			return false, err
		}

		// Record status history for each disbursement waiting for top up using workerpool pattern
		if s.statusHistoryWP == nil {
			s.newStatusHistoryWP()
		}

		wg := sync.WaitGroup{}
		for _, disbursementID := range disbursementIDs {
			wg.Add(1)
			_ = s.statusHistoryWP.Invoke(&statusHistoryWPData{
				ctx:            ctx,
				wg:             &wg,
				disbursementId: disbursementID,
				actor:          constant.UserSystemType,
				statusType:     constant.DisbursementStatusHistoryWaitingForTopUp,
			})
		}
		wg.Wait()

		if request.BulkID != "" {
			// Update bulk disbursement status to PENDING
			if err := s.disbursementRepo.UpdateBulkDisbursementStatusByID(trxCtx, request.BulkID, constant.BulkDisbursementStatusPending); err != nil {
				return false, err
			}

			sampleDisbursement, err := s.disbursementRepo.FindByID(trxCtx, disbursementIDs[0])
			if err != nil {
				return false, err

			} else if sampleDisbursement == nil {
				return false, pkgErrs.New(response.HttpErrRequest, constant.ErrDisbursementNotFound)
			}

			// Send callback only when created from OPEN API
			if sampleDisbursement.CreatedFrom != nil && *sampleDisbursement.CreatedFrom == constant.DisbursementCreatedFromOpenApi {
				if err = s.sendCallback(ctx, request.BulkID, sampleDisbursement.MerchantID, constant.BulkDisbursementStatusPending, constant.CallbackEventPayoutPending); err != nil {
					return false, err
				}
			}
		}
		return false, nil
	}
	return true, nil
}
