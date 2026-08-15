package disbursementService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"golang.org/x/sync/errgroup"
)

func (s *DisbursementService) ApprovalAction(ctx context.Context, request *disbursementModel.ApprovalActionsRequest) (resp *disbursementModel.ApprovalActionsResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ApprovalAction")
	defer segment.End()

	var (
		isApprovalComplete = false
		approvedList       = make([]string, len(request.Approve))
		actionSummary      = &disbursementModel.ActionTransactionSummary{}
	)
	for i, d := range request.Approve {
		approvedList[i] = d.DisbursementID
	}
	if len(approvedList) > 0 {

		if merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantID); err != nil {
			s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
			return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)

		} else if merchant == nil {
			return nil, pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrMerchantNotFound)

		} else if merchant.ParentID.String != "" { // Sub-Merchant
			ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
		}

		actionSummary, err = s.disbursementRepo.GetActionTransactionSummary(ctx, request.MerchantID, approvedList)
		if err != nil {
			s.logger.Error(ctx, "Get action transaction summary", logger.Error(err))
			return nil, pkgErrors.New(httpResponse.HttpErrDatabase, err)

		} else if actionSummary == nil || actionSummary.Total != len(approvedList) {
			return nil, pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrMerchantNotAllowedPerformAction)
		}

		dailyLimit, err := s.self.ValidateDailyTransactionLimit(ctx, request.MerchantID, actionSummary.TotalAmount)
		if err != nil {
			return nil, err
		}
		defer func() { _ = dailyLimit.Close(context.WithoutCancel(ctx), isApprovalComplete) }()
	}

	var (
		now                = time.Now()
		resultRejectedChan = make(chan string, 1)
		errG               = new(errgroup.Group)
		response           disbursementModel.ApprovalActionsResponse
	)

	defer monitor.WriteAndSend(
		ctx, "disbursement-approval-action", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("merchant_id:%s", request.MerchantID),
				fmt.Sprintf("bulk_id:%s", request.BulkID),
				fmt.Sprintf("approved:%d", len(request.Approve)),
				fmt.Sprintf("rejected:%d", len(request.Reject)),
				"proc_identifier:approval-action",
			}
		},
	)

	// Approve
	errG.Go(func() error {
		requestApprove := &disbursementModel.ApproveRequest{
			ApproveAction: request.Approve,
			MerchantID:    request.MerchantID,
			ApprovedBy:    request.UserID,
			BulkID:        request.BulkID,
			TotalAmount:   actionSummary.TotalAmount,
			CreatedFrom:   constant.DisbursementCreatedFromMerchantPortal,
			IsCompleted:   false,
		}

		errProcess := s.self.Approve(ctx, requestApprove)
		isApprovalComplete = requestApprove.IsCompleted // Status update on approval process

		return errProcess
	})

	// Reject
	errG.Go(func() error {
		requestReject := &disbursementModel.RejectRequest{
			RejectAction: request.Reject,
			MerchantID:   request.MerchantID,
			RejectedBy:   request.UserID,
			BulkID:       request.BulkID,
		}
		rejectedFile, err := s.Reject(ctx, requestReject)

		resultRejectedChan <- rejectedFile

		return err
	})

	err = errG.Wait()

	close(resultRejectedChan)
	if err != nil {
		approval, ok := err.(*disbursementModel.ApprovalResultErr)
		if !ok {
			return nil, err
		}

		// currently we just validate the beneficiary limitation
		response.BeneficiaryLimitExceeded = approval.BeneficiaryLimitExceeded
	}
	isApprovalComplete = true

	// Update bulk disbursement status to DONE when all rejected
	if request.BulkID != "" && len(request.Approve) == 0 && len(request.Reject) > 0 {
		if err := s.disbursementRepo.UpdateBulkDisbursementStatusByID(ctx, request.BulkID, constant.BulkDisbursementStatusDone); err != nil {
			return nil, err
		}
	}

	response.FileRejected = <-resultRejectedChan
	return &response, nil
}

func (s *DisbursementService) ValidateBatchPayoutItems(ctx context.Context, request *disbursementModel.ApprovalActionsRequest) (*disbursementModel.ApprovalActionsRequest, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateBatchPayoutItems")
	defer segment.End()

	if request.BulkID == "" {
		return request, nil
	}

	payouts, err := s.disbursementRepo.GetAllDisbursementByBulkID(ctx, request.BulkID)
	if err != nil {
		s.logger.Error(ctx, "failed to get bulk disbursement items", logger.Error(err), logger.String("bulkID", request.BulkID))
		return nil, err
	}

	if len(payouts) != (len(request.Approve) + len(request.Reject)) {
		s.logger.Warn(ctx, "batch payout items did not match",
			logger.String("bulkID", request.BulkID),
			logger.Int("batchItems", len(payouts)),
			logger.Int("payoutRequestItems", len(request.Approve)+len(request.Reject)),
		)
		return nil, pkgErrors.New(httpResponse.HttpErrUnprocessableContent, constant.ErrInvalidBatchPayoutItem)
	}

	validPayouts := make(map[string]bool, len(payouts))
	for _, payout := range payouts {
		validPayouts[payout.UUID] = true
	}

	validApproveItem := []disbursementModel.ApproveActionObject{}
	for _, approveItem := range request.Approve {
		if _, ok := validPayouts[approveItem.DisbursementID]; ok {
			validApproveItem = append(validApproveItem, approveItem)
			continue
		}

		s.logger.Info(ctx, "approval payout item is not related to the batch",
			logger.String("payoutID", approveItem.DisbursementID),
			logger.String("bulkID", request.BulkID),
		)
		return nil, pkgErrors.New(httpResponse.HttpErrUnprocessableContent, constant.ErrInvalidBatchPayoutItem)
	}

	validRejectItem := []disbursementModel.RejectActionObject{}
	for _, rejectItem := range request.Reject {
		if _, ok := validPayouts[rejectItem.DisbursementID]; ok {
			validRejectItem = append(validRejectItem, rejectItem)
			continue
		}

		s.logger.Info(ctx, "rejected payout item is not related to the batch",
			logger.String("payoutID", rejectItem.DisbursementID),
			logger.String("bulkID", request.BulkID),
		)
		return nil, pkgErrors.New(httpResponse.HttpErrUnprocessableContent, constant.ErrInvalidBatchPayoutItem)
	}

	request.Approve = validApproveItem
	request.Reject = validRejectItem

	return request, nil
}
