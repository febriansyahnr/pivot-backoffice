package dukcapilgatewayrepository

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const urlVerify = "/CALL_VERIFY_BY_ELEMEN"

func (r *DukcapilGatewayRepository) VerifyIdentity(ctx context.Context, req *dukcapilmodel.VerifyRequest) (*dukcapilmodel.VerifyResult, error) {
	ctx, segment := otelTracer.Start(ctx, "/internal/repository/dukcapilGateway/verify")
	defer segment.End()

	request := dukcapilmodel.ToGatewayRequest(req)
	request.UserID = r.secret.Dukcapil.UserID
	request.Password = r.secret.Dukcapil.Password
	request.IPUser = r.secret.Dukcapil.IP
	request.Threshold = constant.GetDukcapilMinimumThreshold(r.config.Environment, &r.config.Dukcapil, r.logger)

	url := r.config.Dukcapil.URL + urlVerify
	response, httpCode, err := r.httpRequest.POST(ctx, url, request, map[string]string{})
	if err != nil {
		r.logger.Error(ctx, "error when call dukcapil gateway", logger.Error(err), logger.Any("httpCode", httpCode), logger.Any("request", request), logger.Any("response", string(response)))
		return nil, err
	}

	var responseData dukcapilmodel.GatewayVerifyResponse
	err = json.Unmarshal(response, &responseData)
	if err != nil {
		r.logger.Error(ctx, "error when unmarshal dukcapil response", logger.Error(err), logger.Any("response", string(response)))
		return nil, err
	}
	if len(responseData.Content) == 0 {
		r.logger.Warn(ctx, "empty response content from dukcapil gateway", logger.Any("response", string(response)))
		return nil, constant.ErrDukcapilInvalidIdentity
	}
	if constant.IncorrectSetupResponseCode[responseData.Content[0].ResponseCode] {
		r.logger.Error(ctx, "incorrect in setup dukcapil gateway variable.", logger.Any("response", string(response)))
		return nil, constant.ErrSetupDukcapilConfig
	}
	// Return response data on invalid identity so that proof can be saved into DB
	if constant.FailedResponseCode[responseData.Content[0].ResponseCode] {
		r.logger.Info(ctx, "invalid dukcapil identity")
		return &responseData.Content[0], constant.ErrDukcapilInvalidIdentity
	}

	return &responseData.Content[0], nil
}
