package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) CreateEWalletPaymentLink(ctx context.Context, request *ewallet.EwalletPaymentRequest) (*ewallet.EwalletPaymentLinkResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/CreateEWalletPaymentLink")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/debit/payment-host-to-host", r.config.SnapCoreConfig.BaseUrl)
	headers := map[string]string{
		constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
	}
	if r.config.Environment != constant.EnvironmentProduction {
		headers = overridePaymentSimulationForHeader(ctx, headers)
	}

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		headers,
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request to create ewallet payment link", logger.Error(err))
		return nil, err
	}

	var resp ewallet.SnapCoreEwalletPaymentLinkResponse

	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read ewallet payment link response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "Response from create ewallet payment link", logger.Any("request", request), logger.Any("response", map[string]any{
		"status": statusCode, "body": string(response),
	}))

	errMsg := resp.Message
	if errMsg == "" {
		errMsg = constant.ErrPaymentPartnerInGeneral.Error()
	}
	if resp.Error != nil && resp.Error.Message != "" {
		errMsg = resp.Error.Message
	}

	if errType, mapped := mapPartnerHTTPStatusToErrorType(statusCode); mapped {
		return nil, pkgErrors.New(errType, errors.New(errMsg))
	}

	return resp.Data, nil

}

func (r *snapCoreRepository) InquiryStatusEWalletPayment(ctx context.Context, request *ewallet.EWalletInquiryStatusRequest) (*ewallet.EWalletInquiryStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/InquiryStatusEWalletPayment")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/debit/%s/status", r.config.SnapCoreConfig.BaseUrl, request.TransactionID)

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		nil,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request to inquiry status ewallet payment", logger.Error(err))
		return nil, err
	}

	var resp ewallet.EWalletInquiryStatusResponse

	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read inquiry status ewallet payment response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "Response from inquiry status ewallet payment", logger.Any("request", request), logger.Any("response", map[string]any{
		"status": statusCode, "body": string(response),
	}))

	if errType, mapped := mapPartnerHTTPStatusToErrorType(statusCode); mapped {
		return nil, pkgErrors.New(errType, constant.ErrPaymentPartnerInGeneral)
	}

	return &resp, nil

}
