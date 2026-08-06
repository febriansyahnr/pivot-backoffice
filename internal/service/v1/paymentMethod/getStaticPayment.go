package paymentMethodService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) GetStaticQRPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) (*paymentModel.PaymentMethodWithPivot, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/GetPaymentMethodByMerchant")
	defer segment.End()

	filter.Category = constant.ProductPayment
	filter.Status = constant.StatusActive
	filter.Type = constant.ChannelQris

	paymentMethod, err := s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method merchant", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentMethod == nil {
		s.logger.Error(ctx, "payment method merchant not found")
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	result, err := s.paymentRepo.FilterStaticQrisList(ctx, paymentModel.StaticQrisFilterRequest{
		MerchantID:      filter.MerchantID,
		Status:          constant.StatusActive,
		PaymentMethodID: paymentMethod.UUID,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get list QR Static")
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	qrPayments := []paymentModel.StaticQRPaymentItem{}
	for _, payment := range result.Data.([]paymentModel.StaticQrisListResponse) {
		qrPayments = append(qrPayments, paymentModel.StaticQRPaymentItem{
			MerchantID:               payment.MerchantID,
			PaymentSessionID:         payment.UUID,
			PaymentClientReferenceID: payment.ReferenceID,
			StoreID:                  payment.StoreID,
			IsDerived:                payment.MerchantID != filter.MerchantID,
			CreatedAt:                payment.CreatedAt,
			ExpiredAt:                util.ValueOfPtr(payment.ExpiredAt),
		})
	}

	if len(qrPayments) > 0 {
		paymentMethod.QRPayments = qrPayments
	}

	return paymentMethod, nil
}
