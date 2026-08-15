package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) CreateVirtualAccountConfig(
	ctx context.Context,
	request *snapCoreModel.CreateVirtualAccountConfigRequest) (*snapCoreModel.VirtualAccountConfigResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/CreateVirtualAccountConfig")
	defer segment.End()

	builtUrl := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/config", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		builtUrl,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request create virtual account config", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.VirtualAccountConfigResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read create virtual account config response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when create virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when create virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "response from create virtual account config", logger.String("body", string(response)))
	return resp.Data, nil
}

func (r *snapCoreRepository) GetVirtualAccountConfig(
	ctx context.Context,
	request *snapCoreModel.GetVirtualAccountConfigRequest) ([]*snapCoreModel.VirtualAccountConfigResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/GetVirtualAccountConfig")
	defer segment.End()

	params := url.Values{}
	params.Add("merchant_id", request.MerchantID)
	params.Add("mid", request.MID)
	params.Add("bin_prefix", request.BinPrefix)
	params.Add("type", request.Type)
	params.Add("integration_type", request.IntegrationType)
	params.Add("acquirer", request.Acquirer)
	params.Add("status", request.Status)

	builtUrl := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/config?%s", r.config.SnapCoreConfig.BaseUrl, params.Encode())

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.GET(
		ctx,
		builtUrl,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
			constant.HeaderXMerchantId:         request.MerchantID,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get virtual account config", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.GetVirtualAccountConfigResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get virtual account config response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	// Test hook for setting up metadata in tests
	if r.testVAConfigPostProcessor != nil {
		r.testVAConfigPostProcessor(resp.Data)
	}

	for _, datum := range resp.Data {
		if datum.Metadata.Valid {
			_ = json.Unmarshal(datum.Metadata.JSONText, &datum.MetadataObj)
		}
	}

	r.logger.Info(ctx, "response from get virtual account config", logger.String("body", string(response)))
	return resp.Data, nil
}

func (r *snapCoreRepository) UpdateVirtualAccountConfigPrefix(
	ctx context.Context,
	request *snapCoreModel.UpdateVirtualAccountConfigPrefixRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/CreateVirtualAccountConfig")
	defer segment.End()

	builtUrl := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/config/prefix", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.PATCH(
		ctx,
		builtUrl,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request update virtual account config", logger.Error(err))
		return err
	}

	var resp snapCoreModel.UpdateVirtualAccountConfigPrefixResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read update virtual account config response body", logger.Error(err))
		return err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when update virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	if statusCode >= 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when update virtual account config, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	r.logger.Info(ctx, "response from update virtual account config", logger.String("body", string(response)))
	return nil
}
