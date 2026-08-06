package recurringContractService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) Cancel(ctx context.Context, request model.CancelRecurringContractRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/recurringContract/Cancel")
	defer span.End()

	recurringContract, err := s.repo.GetDetailByID(ctx, request.MerchantID, request.RecurringID)
	if err != nil {
		s.log.Error(ctx, "Failed to retrieve recurring payment contract details", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if recurringContract == nil {
		return constant.NewErrResourceNotFound("recurring payment contract", request.RecurringID)

	} else if recurringContract.Status == constant.RecurringContractStatusInactive {
		return nil
	}

	err = s.repo.UpdateRecurringContractStatus(ctx, request.RecurringID, constant.RecurringContractStatusInactive, request.UpdatedBy)
	if err != nil {
		s.log.Error(ctx, "Failed to update the recurring payment contract status", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return nil
}
