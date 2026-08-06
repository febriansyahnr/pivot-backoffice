package callbackService

import (
	"context"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackPartnerModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) ResendCallback(ctx context.Context, id, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/ResendCallback")
	defer segment.End()

	// Get callback log, include merchant validation.
	callback, err := s.GetCallbackLogDetail(ctx, id, merchantID)
	if err != nil {
		return err
	}

	callbackLog := callback.ToCallbackLog()

	var requestData callbackPartnerModel.CallbackPayloadRequest
	if errUnmarshal := json.Unmarshal([]byte(callbackLog.Request), &requestData); errUnmarshal != nil {
		s.logger.Error(ctx, "Error unmarshalling JSON", logger.Error(errUnmarshal))
		return errUnmarshal
	}

	reqCallback := callbackPartnerModel.CallbackRequest{
		URL:     callback.URL,
		Request: requestData,
	}

	callbackApiKey, err := s.merchantSvc.GetOrGenerateCallbackApiKey(ctx, merchantID)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-API-KEY": *callbackApiKey,
	}

	// Incr retry
	callbackLog.Retry = callbackLog.Retry + 1
	callbackLog.UpdatedAt = time.Now().UTC()

	// Resend callback, this callback only support non-snap CALLBACK
	response, err := s.callbackPartnerSvc.Callback(ctx, reqCallback, headers)
	if err != nil {
		s.logger.Error(ctx, "Got error when resend callback", logger.Error(err))

		// Update failed status and response
		callbackLog.Status = constant.CallbackStatusFailed
		callbackLog.Response = &response
		if errUpdate := s.callbackRepo.UpdateCallbackLog(ctx, callbackLog); errUpdate != nil {
			return errUpdate
		}

		return nil
	}

	// Update success status and response
	callbackLog.Status = constant.CallbackStatusDelivered
	callbackLog.Response = &response
	return s.callbackRepo.UpdateCallbackLog(ctx, callbackLog)
}

func (s *CallbackService) ResendSNAPCallback(ctx context.Context, id, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/ResendSnapCallback")
	defer segment.End()

	// Get callback log, include merchant validation.
	callback, err := s.GetCallbackLogDetail(ctx, id, merchantID)
	if err != nil {
		return err
	}

	callbackLog := callback.ToCallbackLog()

	var requestData callbackPartnerModel.CallbackSNAPPayloadRequest
	if errUnmarshal := json.Unmarshal([]byte(callbackLog.Request), &requestData); errUnmarshal != nil {
		s.logger.Error(ctx, "Error unmarshalling JSON", logger.Error(errUnmarshal))
		return errUnmarshal
	}

	reqCallback := callbackPartnerModel.CallbackRequest{
		URL:     callback.URL,
		Request: requestData.Body,
	}
	header := map[string]string{
		"Content-Type": "application/json",
		"X-TIMESTAMP":  requestData.Header["X-TIMESTAMP"],
		"X-CLIENT-KEY": requestData.Header["X-CLIENT-KEY"],
		"X-SIGNATURE":  requestData.Header["X-SIGNATURE"],
	}

	// Incr retry
	callbackLog.Retry = callbackLog.Retry + 1
	callbackLog.UpdatedAt = time.Now().UTC()

	// Resend callback, this callback only support non-snap CALLBACK
	response, err := s.callbackPartnerSvc.Callback(ctx, reqCallback, header)
	if err != nil {
		s.logger.Error(ctx, "Got error when resend callback", logger.Error(err))

		// Update failed status and response
		callbackLog.Status = constant.CallbackStatusFailed
		callbackLog.Response = &response
		if errUpdate := s.callbackRepo.UpdateCallbackLog(ctx, callbackLog); errUpdate != nil {
			return errUpdate
		}

		return nil
	}

	// Update success status and response
	callbackLog.Status = constant.CallbackStatusDelivered
	callbackLog.Response = &response
	return s.callbackRepo.UpdateCallbackLog(ctx, callbackLog)
}
