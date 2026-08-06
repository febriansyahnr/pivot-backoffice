package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) GetMIDByAcquirerMID(
	ctx context.Context,
	acquirerMID string,
) (*creditcardCoreProcessorModel.MIDResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMIDByAcquirerMID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/acquirer-mid/%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl,
		acquirerMID)
	r.logger.Info(ctx, "GetMIDByAcquirerMID", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid by acquirer mid", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.MIDResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get  mid by acquirer mid body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetMIDByAcquirerMID", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get mid by acquirer mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get mid by acquirer mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) GetMID(
	ctx context.Context,
	midId string,
) (*creditcardCoreProcessorModel.MIDResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl,
		midId)
	r.logger.Info(ctx, "GetMID", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.MIDResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get mid body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetMID", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) ValidateMidInstallmentBins(
	ctx context.Context,
	request *creditcardCoreProcessorModel.ValidateMIDInstallmentBinsRequest,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/installment/validate-bin",
		r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "ValidateMidInstallmentBins", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid", logger.Error(err))
		return err
	}

	var resp creditcardCoreProcessorModel.ValidateMIDInstallmentBinsResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read response body", logger.Error(err))
		return err
	}

	r.logger.Info(ctx, "Validate MID installment bins", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil && resp.Message == "" {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		if statusCode == http.StatusForbidden {
			err = pkgErrors.New(httpResponse.HttpErrForbidden, errors.New(errMsg))
		}
		r.logger.Error(ctx, fmt.Sprintf("got error 4xx when validate mid installment bins, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when validate mid installment bins, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	return nil
}

func (r *creditcardCoreProcessorRepository) CreateMID(
	ctx context.Context,
	request *creditcardCoreProcessorModel.CreateMIDRequest,
) (*creditcardCoreProcessorModel.CreateMIDResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/CreateMID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid",
		r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "CreateMID", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request create mid", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.CreateMIDResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read create mid body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CreateMID", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil && errMsg == "" {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when create mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when create mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) UpdateMID(
	ctx context.Context,
	request *creditcardCoreProcessorModel.UpdateMIDRequest,
) (*creditcardCoreProcessorModel.UpdateMIDResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/UpdateMID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl, request.UUID)
	r.logger.Info(ctx, "UpdateMID", logger.String("url", url))

	response, statusCode, err := r.httpRequest.PUT(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request update mid", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.UpdateMIDResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read update mid body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UpdateMID", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when update mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when update mid, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) CreateMIDMap(
	ctx context.Context,
	request *creditcardCoreProcessorModel.CreateMIDMapRequest,
) (*creditcardCoreProcessorModel.CreateMIDMapResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/CreateMIDMap")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/map",
		r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "CreateMIDMap", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request create mid map", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.CreateMIDMapResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read create mid map body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CreateMIDMap", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when create mid map, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when create mid map, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) GetMIDList(
	ctx context.Context,
	request *creditcardCoreProcessorModel.GetMIDListRequest,
) (*creditcardCoreProcessorModel.MIDListResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMIDList")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/list?limit=%d&page=%d",
		r.config.CreditcardCoreProcessorConfig.BaseUrl, request.Limit, request.Page)
	url += "&mid=" + request.Mid
	url += "&acquirer=" + request.Acquirer
	url += "&name=" + request.Name
	url += "&type=" + request.Type
	url += "&transaction_type=" + request.TransactionType
	url += "&installment_type=" + request.InstallmentType
	if request.IsDefault != nil {
		url += "&is_default=" + strconv.FormatBool(*request.IsDefault)
	}
	if request.IsActive != nil {
		url += "&is_active=" + strconv.FormatBool(*request.IsActive)
	}
	r.logger.Info(ctx, "GetMIDList", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid list", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.MIDListResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get mid list body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetMIDList", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get mid list, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get mid list, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) GetMIDMapList(
	ctx context.Context,
	request *creditcardCoreProcessorModel.GetMIDMapListRequest,
) (*creditcardCoreProcessorModel.MIDMapListResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMIDMapList")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/map/list?limit=%d&page=%d",
		r.config.CreditcardCoreProcessorConfig.BaseUrl, request.Limit, request.Page)
	if request.MerchantId != "" {
		url += "&merchant_id=" + request.MerchantId
	}
	r.logger.Info(ctx, "GetMIDMapList", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid map list", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.MIDMapListResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get mid map list body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetMIDMapList", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get mid map list, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get mid map list, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) UpdateMIDMapPriority(
	ctx context.Context,
	request creditcardCoreProcessorModel.UpdateMIDMapPriorityRequest,
) (*creditcardCoreProcessorModel.UpdateMIDMapResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/UpdateMIDMapPriority")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/map/status/%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl, request.MidMapID)
	r.logger.Info(ctx, "UpdateMIDMapPriority", logger.String("url", url))

	response, statusCode, err := r.httpRequest.PUT(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request update mid map status", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.UpdateMIDMapResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read update mid map status body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UpdateMIDMapPriority", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when update mid map status, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when update mid map status, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *creditcardCoreProcessorRepository) FindMIDMapByMerchant(
	ctx context.Context,
	request *creditcardCoreProcessorModel.FindMIDMapByMerchantRequest,
) (*creditcardCoreProcessorModel.MIDMapResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/GetMIDMapList")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/mid/map/merchant/%s/mid/%s",
		r.config.CreditcardCoreProcessorConfig.BaseUrl, request.MerchantID, request.MidID)
	r.logger.Info(ctx, "FindMIDMapByMerchant", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get mid map detail", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.MIDMapResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get mid map detail body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "FindMIDMapByMerchant", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.(string)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get mid map detail, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get mid map detail, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
