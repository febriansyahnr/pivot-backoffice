package disbursementService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Cancel cancels disbursement transactions for the specified merchant.
//
// This function allows merchants to cancel disbursement transactions that have
// insufficient balance issues. It processes both bulk disbursement IDs and
// individual batch IDs, filtering for disbursements that can be cancelled.
//
// The function performs the following operations:
//   - Validates that at least one ID type (BatchBulkID or BatchID) is provided
//   - Processes bulk IDs by fetching all associated disbursements
//   - Processes individual batch IDs directly
//   - Filters disbursements to only include those with insufficient balance reason
//   - Updates all valid disbursements with cancellation reason
func (s *DisbursementService) Cancel(ctx context.Context, payload *disbursementModel.CancelPayoutRequest) ([]string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/Cancel")
	defer segment.End()

	var (
		validBatchPayoutID   = []string{}
		validOpenApiBatchIDs = map[string]bool{}
	)

	if len(payload.BatchBulkID) == 0 && len(payload.BatchID) == 0 {
		return validBatchPayoutID, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrInvalidDisbursementCancelRequest)
	}

	for _, bulkID := range payload.BatchBulkID {
		batchPayout, err := s.disbursementRepo.GetAllDisbursementByBulkID(ctx, bulkID)
		if err != nil {
			return validBatchPayoutID, pkgErrors.New(response.HttpErrDatabase, err)
		}

		for _, payout := range batchPayout {
			if payout.CreatedFrom != nil && *payout.CreatedFrom == constant.DisbursementCreatedFromOpenApi {
				validOpenApiBatchIDs[bulkID] = true
			}

			// cancel flow only support for insufficient balance
			if util.ValueOfPtr(payout.ReasonType) != constant.DisbursementReasonTypeInsufficientBalance {
				return []string{}, pkgErrors.New(response.HttpErrUnprocessableContent, fmt.Errorf("%s, id: %s", constant.ErrDisbursementCannotBeCancelled, payout.UUID))
			}

			validBatchPayoutID = append(validBatchPayoutID, payout.UUID)
		}

		err = s.disbursementRepo.UpdateBulkDisbursementStatusByID(ctx, bulkID, constant.BulkDisbursementStatusDone)
		if err != nil {
			return []string{}, pkgErrors.New(response.HttpErrDatabase, err)
		}
	}

	payouts, err := s.disbursementRepo.GetByIDs(ctx, payload.BatchID)
	if err != nil {
		return []string{}, pkgErrors.New(response.HttpErrDatabase, err)
	}

	for _, payout := range payouts {
		// cancel flow only support for insufficient balance
		if util.ValueOfPtr(payout.ReasonType) != constant.DisbursementReasonTypeInsufficientBalance {
			return []string{}, pkgErrors.New(response.HttpErrUnprocessableContent, fmt.Errorf("%s, id: %s", constant.ErrDisbursementCannotBeCancelled, payout.UUID))
		}

		validBatchPayoutID = append(validBatchPayoutID, payout.UUID)
	}

	err = s.disbursementRepo.UpdateReasonByIDs(
		ctx,
		validBatchPayoutID,
		constant.DisbursementReasonTypeCancelled,
		"The resource was cancelled by users",
	)
	if err != nil {
		return []string{}, pkgErrors.New(response.HttpErrDatabase, err)
	}

	for _, payoutID := range validBatchPayoutID {
		s.recordDisbursementCancelled(ctx, payoutID, constant.DisbursementReasonTypeCancelled)
	}

	for bulkID, _ := range validOpenApiBatchIDs {
		err = s.sendCallback(ctx, bulkID, payload.MerchantID, constant.DisbursementReasonTypeCancelled, constant.CallbackEventPayoutCancelled)
		if err != nil {
			s.logger.Error(ctx, "Failed to send callback for cancelled payout", logger.Error(err), logger.String("merchantId", payload.MerchantID), logger.String("bulkID", bulkID))
		}
	}

	return validBatchPayoutID, nil
}
