package creditcard

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CreditCardService) Void(
	ctx context.Context,
	request *creditcardModel.VoidRequest,
) (*creditcardModel.VoidResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/Void")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantID.String(),
		From:        serviceName,
	})

	payment, err := c.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID.String(), request.ReferenceID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil {
		err = constant.ErrCreditcardReferenceIdNotFound
		c.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("reference_id", request.ReferenceID))
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	if payment.ProcessorReferenceNumber == nil {
		err = constant.ErrCreditcardPaymentIsUnpaid
		c.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("reference_id", request.ReferenceID), logger.String("status", payment.Status))
		return nil, pkgErrors.New(response.HttpErrForbidden, err)
	}

	switch {
	case payment.Status == constant.StatusFailed:
		status, err := payment.GetCreditcardPaymentStatus()
		if err != nil {
			c.logger.Error(ctx, err.Error(), logger.Error(err))
			return nil, pkgErrors.New(response.HttpErrInternal, err)
		}

		if status == constant.CreditCardStatusVoid {
			err = constant.ErrCreditcardPaymentAlreadyVoid
			c.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("reference_id", request.ReferenceID), logger.String("processor_status", payment.Status))
			return nil, pkgErrors.New(response.HttpErrDupCheck, err)
		}

		err = constant.ErrCreditcardPaymentNotSuccess
		c.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("reference_id", request.ReferenceID), logger.String("status", payment.Status))
		return nil, pkgErrors.New(response.HttpErrForbidden, err)
	case payment.Status != constant.StatusSuccess && payment.Status != constant.UnifiedPaymentSessionStatusPaid:
		err = constant.ErrCreditcardPaymentNotSuccess
		c.logger.Error(ctx, err.Error(), logger.Error(err), logger.String("reference_id", request.ReferenceID), logger.String("status", payment.Status))
		return nil, pkgErrors.New(response.HttpErrForbidden, err)
	}

	// Call to Creditcard Core Processor
	voidResp, err := c.creditcardCoreProcessorRepo.Void(ctx, &creditcardCoreProcessorModel.VoidRequest{
		MerchantID:            payment.MerchantID,
		ClientTransactionID:   *payment.ReferenceID,
		AcquirerTransactionID: *payment.ProcessorReferenceNumber,
	})
	if err != nil {
		c.logger.Error(ctx, "Void - error when void creditcard transaction to creditcard core processor", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, err)
	}

	return &creditcardModel.VoidResponse{
		Status:                voidResp.Status,
		AcquirerTransactionID: voidResp.AcquirerTransactionID,
		GrandTotalAmount:      voidResp.GrandTotalAmount,
		Currency:              voidResp.Currency,
		CardBrand:             voidResp.CardBrand,
		CreatedAt:             voidResp.CreatedAt,
	}, nil
}
