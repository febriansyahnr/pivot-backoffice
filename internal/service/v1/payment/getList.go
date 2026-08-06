package paymentService

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentService) GetListForInternalDashboard(ctx context.Context, request *paymentModel.GetListFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetListForInternalDashboard")
	defer segment.End()

	list, err := s.paymentRepo.GetList(ctx, request)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	return list, nil
}
