package recurringContractService

import (
	"context"
	"errors"
	"strconv"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetRecurringByID retrieves recurring contract details by ID for the merchant dashboard.
// This function is specifically designed for dashboard use only and provides comprehensive
// recurring contract information including plan, billing, trials, and payment method details.
// Note: This endpoint is intended for merchant dashboard viewing purposes only.
func (s *service) GetRecurringByID(ctx context.Context, request recurringContractModel.GetRecurringByIDRequest) (*recurringContractModel.GetRecurringByIDDashboardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/recurringContract/GetRecurringByID")
	defer segment.End()

	detail, err := s.repo.GetDetailByID(ctx, request.MerchantID, request.RecurringID)
	if err != nil {
		s.log.Error(ctx, "error when get recurring contract detail by ID", logger.Error(err))
		return nil, err
	}

	if detail == nil {
		err := pkgErrors.New(response.HttpErrNotFound, errors.New("recurring contract not found"))
		s.log.Error(ctx, "recurring contract not found", logger.String("recurringId", request.RecurringID))
		return nil, err
	}

	responseData := &recurringContractModel.GetRecurringByIDDashboardResponse{
		UUID:              detail.UUID,
		MerchantID:        detail.MerchantID,
		CustomerID:        detail.CustomerID,
		ClientReferenceID: detail.ClientReferenceID,
		Plan:              detail.Plan,
		Trials:            detail.Trials,
		Billing:           detail.Billing,
		Amount: commonModel.Amount{
			Currency: detail.Currency,
			Value:    strconv.FormatFloat(detail.Amount, 'f', 2, 64),
		},
		Status:    detail.Status,
		UpdatedAt: detail.UpdatedAt,
		CreatedAt: detail.CreatedAt,
	}

	return responseData, nil
}
