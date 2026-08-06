package callbackService

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	callbackPartner "github.com/paper-indonesia/pivot-backoffice/pkg/callback"
	callbackPartnerModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) ProcessCallback(
	ctx context.Context, request *callbackModel.ProcessCallbackRequest,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/ConsumeCallback")
	defer segment.End()

	callback, err := s.callbackRepo.FindCallbackByNameAndMerchantID(ctx, request.Name, request.MerchantID)
	if err != nil {
		return err
	}

	if callback == nil {
		s.logger.Info(ctx, constant.ErrCallbackURLNotConfigured.Error(), logger.Any("callbackRequest", request))
		return constant.ErrCallbackURLNotConfigured
	}

	if request.IsSnap {
		callbackRequest := request.Request
		if unifiedResp, ok := request.Request.(*paymentModel.UnifiedPaymentResponse); ok {
			callbackRequest = unifiedResp.ToSnapPayment()
		} else if chargeResp, ok := request.Request.(*unifiedPaymentModel.ChargeResponse); ok {
			callbackRequest = chargeResp.ToSnapPayment()
		} else {
			if constant.IsPaymentVA(request.Event) {
				var paymentResp paymentModel.PaymentResponse
				jsonData, _ := json.Marshal(request.Request)
				json.Unmarshal(jsonData, &paymentResp)
				callbackRequest = paymentResp.ToSnapVAPaymentResponse()
			}
		}

		// Default snap process
		_, err := s.processSnapNew(ctx, callbackRequest, callback, request.Event)
		if err != nil {
			return err
		}

		return nil
	}

	// Generate API Key first
	callbackApiKey, err := s.merchantSvc.GetOrGenerateCallbackApiKey(ctx, request.MerchantID.String())
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-API-KEY": *callbackApiKey,
	}

	// Req callback
	reqCallback := callbackPartnerModel.CallbackRequest{
		URL: callback.URL,
		Request: callbackPartnerModel.CallbackPayloadRequest{
			Event: request.Event,
			Data:  request.Request,
		},
	}

	requestByte, err := json.Marshal(reqCallback.Request)
	if err != nil {
		s.logger.Error(ctx, "error when marshal request", logger.Error(err))
		return err
	}

	// Store callback logs before call to client service.
	callbackLogs := &callbackModel.CallbackLog{
		UUID:        uuid.New(),
		CallbackID:  callback.UUID,
		Event:       &request.Event,
		Request:     string(requestByte),
		Response:    nil,
		Status:      constant.CallbackStatusPending,
		Retry:       0,
		ReferenceId: callbackModel.GetReferenceId(request.Event, request.Request),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err = s.callbackRepo.CreateCallbackLog(ctx, callbackLogs); err != nil {
		return err
	}

	response, err := s.callbackPartnerSvc.Callback(ctx, reqCallback, headers)
	if err != nil {
		// Update failed status and response
		callbackLogs.Status = constant.CallbackStatusFailed
		callbackLogs.Response = &response
		if errUpdate := s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs); errUpdate != nil {
			return errUpdate
		}

		return err
	}

	// Update success status and response
	callbackLogs.Status = constant.CallbackStatusDelivered
	callbackLogs.Response = &response
	return s.callbackRepo.UpdateCallbackLog(ctx, callbackLogs)
}

func (s *CallbackService) SendMerchantCallback(ctx context.Context, request callbackModel.SendMerchantCallbackRequest) (*callbackModel.SendMerchantCallbackResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/SendMerchantCallback")
	defer segment.End()

	var headers map[string]string

	callbackDetails := &callbackModel.Callback{
		UUID:       util.ParseUUID(request.CallbackId),
		MerchantID: util.ParseUUID(request.MerchantId),
		URL:        request.CallbackUrl,
	}

	if request.IsSnap {
		snapHeaders, err := s.buildSnapRequestHeaders(ctx, request.EventName, callbackDetails, request.RawPayload)
		if err != nil {
			return nil, err
		}
		headers = snapHeaders.ToArray()

	} else {
		merchantApiKey, err := s.merchantSvc.GetOrGenerateCallbackApiKey(ctx, request.MerchantId)
		if err != nil {
			return nil, err
		}
		headers = map[string]string{
			"X-API-KEY": *merchantApiKey,
		}
	}

	callbackRequest := callbackPartnerModel.CallbackRequest{
		URL:     request.CallbackUrl,
		Request: request.RawPayload,
	}
	responseString, err := s.callbackPartnerSvc.Callback(ctx, callbackRequest, headers)
	if err != nil && responseString == "" {
		return nil, err
	}

	statusCode := http.StatusOK

	if err != nil {
		if errHttpClient, ok := err.(*callbackPartner.ErrHttpClient); ok {
			statusCode = errHttpClient.StatusCode()
		}
	}

	callbackResult := &callbackModel.SendMerchantCallbackResponse{
		StatusCode:   statusCode,
		ResponseBody: []byte(responseString),
	}
	if request.IsSnap {
		callbackResult.AdditionalInfo = &callbackModel.SendMerchantCallbackAdditionalInfo{
			Headers: headers,
			URL:     request.CallbackUrl,
		}
	}
	return callbackResult, nil
}

func (s *CallbackService) WriteCallbackLogFromWorkflowTask(ctx context.Context, log callbackModel.WorkflowWriteLogRequest) (*callbackModel.WorkflowWriteLogResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/WriteCallbackLogFromWorkflowTask")
	defer segment.End()

	var requestBytes []byte = log.RawPayload

	if log.IsSnap && log.Response.AdditionalInfo != nil {
		requestBytes, _ = json.Marshal(map[string]any{
			"url":    log.Response.AdditionalInfo.URL,
			"header": log.Response.AdditionalInfo.Headers,
			"body":   log.RawPayload,
		})
	}

	var (
		id, _            = uuid.NewV7()
		status           = constant.CallbackStatusDelivered
		responseBytes, _ = json.Marshal(log.Response.ResponseBody)
	)

	if log.Response.StatusCode >= 400 || log.Response.StatusCode == 0 {
		status = constant.CallbackStatusFailed
	}

	callbackLog := &callbackModel.CallbackLog{
		UUID:        id,
		CallbackID:  util.ParseUUID(log.CallbackId),
		Event:       &log.EventName,
		Request:     string(requestBytes),
		Response:    util.ValueToPtr(string(responseBytes)),
		Status:      status,
		Retry:       log.RetryCount,
		ReferenceId: log.ReferenceId,
		Metadata: &callbackModel.CallbackLogMetadata{
			WorkflowId: log.WorkflowId,
		},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := s.callbackRepo.CreateCallbackLog(ctx, callbackLog); err != nil {
		return nil, fmt.Errorf("write callback log: %v", err)
	}
	return &callbackModel.WorkflowWriteLogResponse{
		CallbackLogId: callbackLog.UUID.String(),
	}, nil
}
