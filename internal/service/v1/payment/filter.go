package paymentService

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FilterPaymentHistory return list of merchant payment history
func (s *PaymentService) FilterPaymentHistory(ctx context.Context, opt paymentModel.FilterPaymentHistoryOption) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/FilterPaymentHistory")
	defer segment.End()

	err := opt.Validate()
	if err != nil {
		return nil, errPkg.New(response.HttpErrRequest, err)
	}

	return s.paymentRepo.FilterPaymentHistory(ctx, opt)
}
