package disbursementService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ResendDisbursementCallback resends the callback for a disbursement transaction
func (s *DisbursementService) ResendDisbursementCallback(ctx context.Context, request *callbackModel.ResendCallbackRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ResendDisbursementCallback")
	defer segment.End()

	var (
		disbursement *disbursementModel.DisbursementWithTransaction
		err          error
	)

	if request.ReferenceID != "" {
		// Load disbursement by ID
		disbursement, err = s.disbursementRepo.FindByID(ctx, request.ReferenceID)
		if err != nil {
			s.logger.Error(ctx, "[ResendDisbursementCallback] Failed to get disbursement by ID", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	} else {
		// Load disbursement by client reference id
		disbursement, err = s.disbursementRepo.FindByMerchantAndReference(ctx, request.MerchantID, request.ClientReferenceID)
		if err != nil {
			s.logger.Error(ctx, "[ResendDisbursementCallback] Failed to get disbursement by client reference id", logger.Error(err))
			return pkgErrs.New(response.HttpErrDatabase, err)
		}
	}

	if disbursement == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound)
	}

	if disbursement.MerchantID != request.MerchantID {
		s.logger.Warn(ctx, "[ResendDisbursementCallback] Merchant ID does not match payment ID", logger.String("MerchantID", disbursement.MerchantID))
		return pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch)
	}

	// Validate that this disbursement was created from OPEN_API
	if disbursement.CreatedFrom == nil || *disbursement.CreatedFrom != constant.DisbursementCreatedFromOpenApi {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("disbursement was not created from OPEN_API"))
	}

	// Validate that the disbursement has a bulk ID
	if disbursement.BulkID == nil || *disbursement.BulkID == "" {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("disbursement does not have a bulk ID"))
	}

	// Load bulk disbursement
	bulkDisbursement, err := s.disbursementRepo.FindBulkDisbursementByID(ctx, *disbursement.BulkID)
	if err != nil {
		s.logger.Error(ctx, "[ResendDisbursementCallback] Failed to get bulk disbursement", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	if bulkDisbursement == nil {
		return pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("bulk disbursement not found"))
	}

	// Validate that the bulk disbursement status is DONE
	if bulkDisbursement.Status != constant.BulkDisbursementStatusDone {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("bulk disbursement status is not DONE (current: %s)", bulkDisbursement.Status))
	}

	// Call existing callback function
	if err := s.sendCallback(ctx, *disbursement.BulkID, disbursement.MerchantID, constant.BulkDisbursementStatusDone, constant.CallbackEventPayoutDone); err != nil {
		s.logger.Error(ctx, "[ResendDisbursementCallback] Failed to send callback", logger.Error(err))
		return pkgErrs.New(response.HttpErrInternal, err)
	}

	s.logger.Info(ctx, "[ResendDisbursementCallback] Successfully resent callback for disbursement", logger.Any("request", request))

	return nil
}
