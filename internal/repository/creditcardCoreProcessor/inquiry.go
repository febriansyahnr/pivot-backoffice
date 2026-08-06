package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) InquiryTransaction(ctx context.Context, payload *creditcardModel.InquiryTransactionRequest) (*creditcardModel.PaymentNotificationDataRequest, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/InquiryTransaction")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/credit-card/check-transaction-status?is_unified_payment=true&check_by_order_status=true&client_transaction_id=%s&acquirer_transaction_id=%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl,
		payload.ClientReferenceID,
		payload.ProcessorReferenceID,
	)
	r.logger.Info(ctx, "inquiring transaction", logger.String("processorRefID", payload.ProcessorReferenceID), logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXMerchantID:         payload.MerchantID,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request inquiry creditcard transaction", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "inquiry response", logger.ByteString("response", response))

	var resp creditcardCoreProcessorModel.GenericApiResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get creditcard transaction body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get list creditcard transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get list creditcard transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	result, err := util.ConvertToStruct[*creditcardModel.PaymentNotificationDataRequest](resp.Data)
	if err != nil {
		r.logger.Error(ctx, "failed to unmarshal inquiry response", logger.Error(err), logger.String("processorRefID", payload.ProcessorReferenceID))
	}

	return result, nil
}
