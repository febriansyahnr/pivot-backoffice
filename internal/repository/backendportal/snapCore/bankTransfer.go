package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	generalSnapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

const (
	errInvalidBankTransferRespFormat = "Invalid bank transfer response format. Must be in valid JSON format."
)

func (r *snapCoreRepository) BankTransfer(
	ctx context.Context, request *snapCoreModel.BankTransferRequest, headerRequest *snapCoreModel.BankTransferHeaderRequest,
) (*snapCoreModel.BankTransferResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/BankTransfer")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/bank-transfer", r.config.SnapCoreConfig.BaseUrl)

	r.logger.Info(ctx, "SnapCoreProcessor - BankTransfer Request", pdkLog.String("url", url), pdkLog.Any("request", request))

	ctx = context.WithValue(ctx, constant.CtxEntryPoint, constant.SnapCoreProcessor)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			"X-EXTERNAL-ID":                    headerRequest.ExternalId,
			"X-MERCHANT-ID":                    headerRequest.MerchantId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request bank transfer", pdkLog.Error(err))
		return nil, err
	}

	var resp snapCoreModel.BankTransferResponse
	if err = json.Unmarshal(response, &resp); err != nil {

		r.logger.Error(
			ctx, errInvalidBankTransferRespFormat,
			pdkLog.Error(err), pdkLog.String("transactionId", headerRequest.ExternalId),
			pdkLog.String("merchantId", headerRequest.MerchantId), pdkLog.Any("request", request), pdkLog.ByteString("response", response),
		)

		// For further investigation process, this transaction will be made PENDING status.
		return &snapCoreModel.BankTransferResponseData{
			ResponseMessage: errInvalidBankTransferRespFormat,
			UUID:            constant.EmptyUUID,
			Status:          constant.SnapCoreBankTransferStatusPending,
			ExternalID:      constant.EmptyUUID,
		}, errors.New(errInvalidBankTransferRespFormat)
	}

	r.logger.Info(ctx, "SnapCoreProcessor - BankTransfer Response", pdkLog.String("response", string(response)))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when do request bank account, errorCode %s", resp.Code), pdkLog.Error(err))
		return &resp.Data, err

	} else if statusCode == http.StatusConflict {
		r.logger.Warn(
			ctx, "Double disbursement indication found, bank transfer response 409 Conflict", pdkLog.Any("request", request), pdkLog.Any("response", resp),
		)
		return nil, constant.ErrDoubleDisbursementIndication

	} else if statusCode >= 400 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when do request bank account, errorCode %s", resp.Code), pdkLog.Error(err))
		return &resp.Data, err
	}
	return &resp.Data, nil
}

func (r *snapCoreRepository) FindBankTransferByExternalID(ctx context.Context, externalId string, forceFailed bool) (*snapCoreModel.BankTransferResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/FindBankTransferByExternalID")
	defer segment.End()

	// Build URL with query params
	baseURL := fmt.Sprintf("%s/api/v1.0/internal/bank-transfer/by-external-id/%s", r.config.SnapCoreConfig.BaseUrl, externalId)
	u, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error(ctx, "error parsing URL", pdkLog.Error(err))
		return nil, err
	}
	// Extract forceFailed from context (default to false)
	isFromRetry := false
	if val := ctx.Value(constant.CtxFromRetry); val != nil {
		isFromRetry, _ = val.(bool)
	}

	query := u.Query()
	if forceFailed {
		query.Set("forceFailed", "true")
	}
	if isFromRetry {
		query.Set("fromRetry", "true")
	}
	u.RawQuery = query.Encode()

	finalURL := u.String()

	r.logger.Info(ctx, "FindBankTransferByExternalID",
		pdkLog.String("url", finalURL),
		pdkLog.String("externalId", externalId),
		pdkLog.Bool("forceFailed", forceFailed))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		finalURL,
		map[string]string{
			"X-Internal-Service-Key": r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get bank transfer by external id", pdkLog.Error(err))
		return nil, err
	}

	var resp snapCoreModel.BankTransferResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get bank transfer by external id response body", pdkLog.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "FindBankTransferByExternalID Response", pdkLog.String("response", string(response)))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get bank transfer by external id, errorCode %s", resp.Code), pdkLog.Error(err))
		return &resp.Data, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get bank transfer by external id, errorCode %s", resp.Code), pdkLog.Error(err))
		return &resp.Data, err
	}

	return &resp.Data, nil
}

// UpdateBankTransferStatus updates the status of a bank transfer in the SnapCore system.
// It sends a POST request to the SnapCore API with the provided request data and handles the response.
func (r *snapCoreRepository) UpdateBankTransferStatus(ctx context.Context, req snapCoreModel.UpdateBankTransferStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/UpdateBankTransferStatus")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/bank-transfer/update-status", r.config.SnapCoreConfig.BaseUrl)
	r.logger.Info(ctx, "FindBankTransferByExternalID", pdkLog.String("url", url), pdkLog.Any("request", req))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		req,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do update bank transfer", pdkLog.Error(err))
		return err
	}

	r.logger.Info(ctx, "update bank transfer Response", pdkLog.String("response", string(response)))

	var resp generalSnapCoreModel.StandardResponse

	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when unmarshal update response", pdkLog.Error(err))
	}

	if statusCode >= 500 {
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when update bank transfer, errorCode %s", resp.Code), pdkLog.Error(err))
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(resp.Message))
		return err
	} else if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when update bank transfer, errorCode %s", resp.Code), pdkLog.Error(err))
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Message))
		return err
	}

	r.logger.Info(ctx, "bank transfer updated", pdkLog.String("disbursementID", req.ExternalID))
	return nil
}

func (r *snapCoreRepository) CheckStatusByExternalId(ctx context.Context, externalId string, checkBankStatement bool) (*snapCoreModel.BankTransferCheckStatusResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/FindBankTransferByExternalID")
	defer segment.End()

	// Build URL with query params
	baseURL := fmt.Sprintf("%s/api/v1.0/internal/bank-transfer/check-status/%s", r.config.SnapCoreConfig.BaseUrl, externalId)
	u, err := url.Parse(baseURL)
	if err != nil {
		r.logger.Error(ctx, "error parsing URL", pdkLog.Error(err))
		return nil, err
	}

	query := u.Query()
	if checkBankStatement {
		query.Set("checkBankStatement", "true")
	}
	u.RawQuery = query.Encode()

	finalURL := u.String()

	r.logger.Info(ctx, "CheckStatusByExternalId", pdkLog.String("url", finalURL))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		finalURL,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request check status bank transfer by external id", pdkLog.Error(err))
		return nil, err
	}

	var resp snapCoreModel.BankTransferCheckStatusResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read check status bank transfer by external id response body", pdkLog.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CheckStatusByExternalId Response", pdkLog.String("response", string(response)))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get bank transfer by external id, errorCode %s", resp.Code), pdkLog.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get bank transfer by external id, errorCode %s", resp.Code), pdkLog.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
