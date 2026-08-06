package recurringContractService

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) UpdateRecurringPayment(ctx context.Context, request model.UpdateRecurringPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/recurringContract/UpdateRecurringPayment")
	defer span.End()

	recurringContract, err := s.repo.GetDetailByID(ctx, request.MerchantID, request.RecurringID)
	if err != nil {
		s.log.Error(ctx, "Failed to retrieve recurring payment contract details", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if recurringContract == nil {
		return constant.NewErrResourceNotFound("recurring payment contract", request.RecurringID)
	}

	dataToBeUpdated := model.UpdateRecurringContractRequest{
		RecurringID:       request.RecurringID,
		BillingCycleCount: request.RecurringPayment.BillingCycle.Count,
	}
	if request.RecurringPayment.InitiateFirstAuthorization {
		dataToBeUpdated.UpdatedAt = time.Now().UTC()
		dataToBeUpdated.UpdatedBy = request.UpdatedBy
		dataToBeUpdated.TransactionID = request.TransactionID
		dataToBeUpdated.PaymentTokenID = request.PaymentTokenID
		dataToBeUpdated.PaymentMethodID = request.PaymentMethodID
	}
	if recurringContract.IsFirstAuthorization() {
		dataToBeUpdated.ActivatedAt = time.Now().UTC()
		dataToBeUpdated.Status = constant.RecurringContractStatusActive
	}

	if err = s.repo.UpdateRecurringContract(ctx, dataToBeUpdated); err != nil {
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return pkgErrs.New(response.HttpErrNotFound, err)
		}
		s.log.Error(ctx, "Failed to update the recurring payment contract", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return nil
}
