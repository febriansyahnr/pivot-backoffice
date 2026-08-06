package callbackWorker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"

	modelSdk "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) SendCallback(ctx context.Context, task *conductor.Task) (*conductor.TaskResult, error) {
	ctx, span := otelTracer.Start(ctx, "port/worker/callback/SendCallback")
	defer span.End()

	var (
		start          = time.Now()
		request        model.SendMerchantCallbackRequest
		deliveryResult *model.SendMerchantCallbackResponse
		err            error
	)
	defer func() {
		metricRequest := model.WorkflowRecordMetricRequest{
			MerchantId: request.MerchantId,
			EventName:  request.EventName,
			DurationMs: time.Since(start).Milliseconds(),
		}
		if err != nil {
			metricRequest.ErrorDetail = err.Error()
		}
		if deliveryResult != nil {
			metricRequest.StatusCode = int64(deliveryResult.StatusCode)
		}

		h.recordMetric(ctx, metricRequest)
	}()

	if err = h.jsonBinder.Bind(&request, task.InputData); err != nil {
		h.logger.Error(ctx, "Failed while binding input data to request model", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}
	request.RawPayload, _ = base64.StdEncoding.DecodeString(request.Payload)

	if deliveryResult, err = h.service.SendMerchantCallback(ctx, request); err != nil {
		outputData := model.WorkflowSendCallbackResponse{
			Status: conductor.TaskStatusRetryableError,
		}
		return conductor.NewTaskResultWithOutputAndError(task, outputData, err), nil
	}

	outputData := model.WorkflowSendCallbackResponse{
		StatusCode:     deliveryResult.StatusCode,
		ResponseBody:   map[string]any{},
		AdditionalInfo: deliveryResult.AdditionalInfo,
		Status:         conductor.TaskStatusNonRetryableError,
	}
	_ = json.Unmarshal(deliveryResult.ResponseBody, &outputData.ResponseBody)

	taskResult := conductor.NewTaskResult(task)

	if outputData.StatusCode >= 400 {
		err = constant.ErrInvokeClientWebhook // Used for error details in metrics
		taskResult.Status = outputData.NonSuccessTaskStatus()
		if taskResult.Status == modelSdk.FailedTask {
			outputData.Status = conductor.TaskStatusRetryableError
		}
		taskResult.ReasonForIncompletion = "merchant responded with non-2xx status"

	} else {
		taskResult.Status = modelSdk.CompletedTask
		outputData.Status = conductor.TaskStatusSuccess
	}
	taskResult.OutputData, _ = modelSdk.ConvertToMap(outputData)

	return taskResult, nil
}
