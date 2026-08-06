package paymentMethodService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentMethodService) FindPaymentMethodByIdAndMerchant(ctx context.Context, paymentMethodID, merchantID string) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/FindPaymentMethodByIdAndMerchant")
	defer segment.End()

	paymentMethod, err := s.paymentMethodRepo.FindPaymentMethodByIdAndMerchant(ctx, paymentMethodID, merchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if paymentMethod == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodNotFound)
	}
	return paymentMethod, nil
}
