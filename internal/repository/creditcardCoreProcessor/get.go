package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) GetTransactionList(
	ctx context.Context,
	request *creditcardCoreProcessorModel.GetTransactionListRequest,
) (*creditcardCoreProcessorModel.GetTransactionDataList, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetTransactionList")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/transaction/list?page=%d&limit=%d&date_from=%s&date_to=%s&transaction_type=%s&charge_status=%s&void_status=%s&client_transaction_id=%s&card_fingerprint=%s&issuing_bank=%s&payment_uuid=%s&charge_from=%s&charge_to=%s&refund_from=%s&refund_to=%s&merchant_id=%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl,
		request.Page,
		request.Limit,
		request.DateFrom,
		request.DateTo,
		request.TrxType,
		request.ChargeStatus,
		request.VoidStatus,
		request.ClientTransactionID,
		request.CardFingerprint,
		request.IssuingBank,
		request.PaymentUUID,
		request.ChargeFrom,
		request.ChargeTo,
		request.RefundFrom,
		request.RefundTo,
		request.MerchantID)
	r.logger.Info(ctx, "GetTransactionList", logger.String("url", url))

	headers := map[string]string{
		constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
	}

	if loc, ok := ctx.Value(constant.CtxTimeLocation).(*time.Location); ok && loc != nil {
		headers[constant.HeaderTimezone] = loc.String()
	}

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		headers,
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get list creditcard transaction", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.GetTransactionList
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get list creditcard transaction body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetTransactionList", logger.ByteString("response", response))

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

	return resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) GetBinDetailByBinNumber(ctx context.Context, merchantId, binNumber string) (*creditcardCoreProcessorModel.GetBinDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/creditcardCoreProcessor/GetBinDetailByBinNumber")
	defer segment.End()

	var (
		start   = time.Now()
		url     = r.config.CreditcardCoreProcessorConfig.BaseUrl + "/api/v1/bin/lookup/" + binNumber
		headers = map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		}
		response, statusCode, err = []byte{}, int(0), error(nil)
	)
	defer func() {
		r.logger.Info(ctx, "BIN details request for merchant ID "+merchantId, logger.String("binNumber", binNumber),
			logger.Any(
				"response", map[string]any{
					"responseBody": string(response),
					"statusCode":   statusCode,
				}),
			logger.Int64("durationMs", time.Since(start).Milliseconds()),
		)
	}()

	response, statusCode, err = r.httpRequest.GET(ctx, url, headers)
	if err != nil {
		r.logger.Error(ctx, "Failed while send request get bin detail by bin number", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, err)
	}

	if statusCode == 404 {
		return nil, nil
	}

	result := creditcardCoreProcessorModel.GenericApiResponseGen[creditcardCoreProcessorModel.GetBinDetailResponse]{}
	if err = json.Unmarshal(response, &result); err != nil {
		r.logger.Error(ctx, "Failed while unmarshalling bin response details", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, fmt.Errorf("%v", err))
	}

	if errMsg := errors.New(util.ToTitle(fmt.Sprintf("%v", result.Error))); statusCode >= 500 {
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, errMsg)

	} else if statusCode >= 400 {
		return nil, pkgErrors.New(httpResponse.HttpErrRequest, errMsg)
	}
	return &result.Data, nil
}

func (r *creditcardCoreProcessorRepository) GetCardEncryptionPublicKey(ctx context.Context, merchantID string) ([]byte, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/creditcardCoreProcessor/GetCardEncryptionPublicKey")
	defer segment.End()

	url := r.config.CreditcardCoreProcessorConfig.BaseUrl + "/api/v1/credit-card/auth"
	headers := map[string]string{
		constant.HeaderXMerchantID: merchantID,
		constant.HeaderToken:       base64.StdEncoding.EncodeToString([]byte(r.secret.CreditcardCoreProcessorSecret.EncryptionPublicKeySecret)),
	}

	responseBody, statusCode, err := r.httpRequest.GET(ctx, url, headers)
	if err != nil {
		r.logger.Error(ctx, "Failed while get encryption public key", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, err)
	}

	if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("Failed to get encryption public key, received http code %d", statusCode), logger.ByteString("responseBody", responseBody))
		if statusCode >= 500 {
			return nil, pkgErrors.New(httpResponse.HttpErrThirdParty, fmt.Errorf("%s", responseBody))
		}
		return nil, pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("%s", responseBody))
	}

	result := struct {
		Data string `json:"data"`
	}{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		r.logger.Error(ctx, "Failed to unmarshal response body", logger.Error(err))
		return nil, pkgErrors.New(httpResponse.HttpErrThirdParty, err)
	}

	publicKeyBytes, err := r.decryptCardEncryptionPublicKey(result.Data)
	if err != nil {
		r.logger.Error(ctx, "Failed to decrypt card encryption public key", logger.Error(err))
		return nil, err
	}
	return publicKeyBytes, nil
}

func (r *creditcardCoreProcessorRepository) decryptCardEncryptionPublicKey(encryptedPublicKey string) ([]byte, error) {
	kek, err := base64.StdEncoding.DecodeString(r.secret.CreditcardCoreProcessorSecret.EncryptionPublicKeySecret)
	if err != nil {
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, fmt.Errorf("decode kek: %w", err))
	}

	iv, err := base64.StdEncoding.DecodeString(r.secret.CreditcardCoreProcessorSecret.EncryptionPublicKeyIV)
	if err != nil {
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, fmt.Errorf("decode iv: %w", err))
	}

	publicKeyURLBase64, err := base64.StdEncoding.DecodeString(encryptedPublicKey)
	if err != nil {
		return nil, pkgErrors.New(httpResponse.HttpErrThirdParty, fmt.Errorf("std encoding for decode public key: %w", err))
	}

	publicKeyBytes, err := base64.URLEncoding.DecodeString(string(publicKeyURLBase64))
	if err != nil {
		return nil, pkgErrors.New(httpResponse.HttpErrThirdParty, fmt.Errorf("url encoding for decode public key: %w", err))
	}

	plaintext, err := r.cryptoProvider.DecryptAESCBC(kek, iv, publicKeyBytes)
	if err != nil {
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, err)
	}
	return plaintext, nil
}
