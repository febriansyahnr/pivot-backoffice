package creditcard

import (
	"context"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CreditCardService) BlockCard(ctx context.Context, request *creditcardModel.BlockCardRequest) error {
	ctx, segment := otelTracer.Start(ctx, "service/v1/creditcard/BlockCard")
	defer segment.End()

	blockCardRequest := &creditcardCoreProcessorModel.BlockCardRequest{
		CardUUID:    request.CardUUID,
		IsBlocked:   request.IsBlocked,
		BlockedTo:   request.BlockedTo,
		BlockReason: request.BlockReason,
	}

	err := c.creditcardCoreProcessorRepo.BlockCard(ctx, blockCardRequest)
	if err != nil {
		c.logger.Error(ctx, "Failed to block card", logger.Error(err))
		return err
	}

	return nil
}