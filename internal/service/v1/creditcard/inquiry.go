package creditcard

import (
	"context"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CreditCardService) InquiryTransaction(ctx context.Context, payload *creditcardModel.InquiryTransactionRequest) (*creditcardModel.PaymentNotificationDataRequest, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/InquiryTransaction")
	defer segment.End()

	res, err := c.creditcardCoreProcessorRepo.InquiryTransaction(ctx, payload)
	if err != nil {
		c.logger.Error(ctx, "failed to inquiry payment", logger.Error(err), logger.String("ref_id", payload.ClientReferenceID))
		return nil, err
	}

	return res, nil
}
