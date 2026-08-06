package callbackService

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	callbackPartnerModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkSnapSign "github.com/paper-indonesia/pdk/go/snap/signature"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) processSnapNew(ctx context.Context, request any, callback *callbackModel.Callback, event string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/processSnapNew")
	defer segment.End()

	// Build SNAP Request Headers
	snapHeader, err := s.buildSnapRequestHeaders(ctx, event, callback, request)
	if err != nil {
		return "", err
	}
	requestBytes, _ := json.Marshal(request)

	requestFull := map[string]any{
		"url":    callback.URL,
		"header": snapHeader.ToArray(),
		"body":   request,
	}

	requestByte, _ := json.Marshal(requestFull)

	// Store callback logs before call to client service.
	referenceId := callbackModel.GetReferenceId(event, request)
	callbackLogs := &callbackModel.CallbackLog{
		UUID:        uuid.New(),
		CallbackID:  callback.UUID,
		Event:       &event,
		Request:     string(requestByte),
		Response:    nil,
		Status:      constant.CallbackStatusPending,
		Retry:       0,
		ReferenceId: referenceId,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err = s.callbackRepo.CreateCallbackLog(ctx, callbackLogs); err != nil {
		return "", err
	}

	// send request
	response, err := s.callbackPartnerSvc.Callback(
		ctx,
		callbackPartnerModel.CallbackRequest{
			URL:     callback.URL,
			Request: requestBytes,
		},
		snapHeader.ToArray(),
	)
	if err != nil {
		callbackLogs.Status = constant.CallbackStatusFailed
		callbackLogs.Response = &response
		if errUpdate := s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs); errUpdate != nil {
			return "", errUpdate
		}

		return "", err
	}
	// Update success status and response
	callbackLogs.Status = constant.CallbackStatusDelivered
	callbackLogs.Response = &response
	return response, s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs)
}

func (s *CallbackService) getSnapAccessTokenB2bNew(ctx context.Context, callback *callbackModel.Callback, event string) (*string, *merchantModel.SNAPAccessTokenB2BResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/getSnapAccessTokenB2b")
	defer segment.End()
	var resp merchantModel.SNAPAccessTokenB2BResp
	tokenCacheKey := fmt.Sprintf("backend-portal:snap-access-token:%s:%s", callback.URL, callback.MerchantID.String())
	tokenBodyCacheKey := fmt.Sprintf("backend-portal:snap-access-token-body:%s:%s", callback.URL, callback.MerchantID.String())

	val, err := s.cacheService.Get(ctx, tokenCacheKey).Result()
	cacheBody, _ := s.cacheService.Get(ctx, tokenBodyCacheKey).Result()
	if val != "" && cacheBody != "" && err == nil {
		if err := json.Unmarshal([]byte(cacheBody), &resp); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal cached token body: %w", err)
		}
		return &val, &resp, nil
	}

	privateKey, err := s.merchantSvc.GetSnapPrivateKey(ctx, callback.MerchantID.String())
	if err != nil {
		s.logger.Error(ctx, "error when get snap private key", logger.Error(err))
		return nil, nil, err
	}

	timestamp := util.SnapCompatible(time.Now())

	b2bTokenSignature := pdkSnapSign.NewB2bTokenSignature(
		pdkSnapSign.B2bTokenSignature{
			PrivateKey: privateKey,
			Timestamp:  timestamp,
			ClientID:   callback.MerchantID.String(),
		},
	)

	signature, _ := b2bTokenSignature.Create()

	url := callback.URL

	header := map[string]string{
		"Content-Type": "application/json",
		"X-TIMESTAMP":  timestamp,
		"X-CLIENT-KEY": callback.MerchantID.String(),
		"X-SIGNATURE":  *signature,
	}

	request := callbackPartnerModel.CallbackRequest{
		URL: url,
		Request: map[string]any{
			"grantType": "client_credentials",
		},
	}

	requestFull := map[string]any{
		"url":    url,
		"header": header,
		"body":   request.Request,
	}

	requestByte, _ := json.Marshal(requestFull)

	// Store callback logs before call to client service.
	// For access token B2B callbacks, there's no reference ID to extract
	callbackLogs := &callbackModel.CallbackLog{
		UUID:        uuid.New(),
		CallbackID:  callback.UUID,
		Event:       &event,
		Request:     string(requestByte),
		Response:    nil,
		Status:      constant.CallbackStatusPending,
		Retry:       0,
		ReferenceId: nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err = s.callbackRepo.CreateCallbackLog(ctx, callbackLogs); err != nil {
		return nil, nil, err
	}

	response, err := s.callbackPartnerSvc.Callback(
		ctx,
		request,
		header,
	)
	if err != nil {

		callbackLogs.Status = constant.CallbackStatusFailed
		callbackLogs.Response = &response
		if errUpdate := s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs); errUpdate != nil {
			return nil, nil, errUpdate
		}
		s.logger.Error(ctx, "error when get snap access token b2b", logger.Error(err))
		return nil, nil, err
	}

	json.Unmarshal([]byte(response), &resp)

	token := resp.AccessToken

	s.cacheService.Set(ctx, tokenCacheKey, token, 14*time.Minute).Err()
	s.cacheService.Set(ctx, tokenBodyCacheKey, response, 14*time.Minute).Err()

	callbackLogs.Status = constant.CallbackStatusDelivered
	callbackLogs.Response = &response
	if err = s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs); err != nil {
		return nil, nil, err
	}

	return &token, &resp, nil
}

func getUrlPath(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	return parsedUrl.Path, nil
}
