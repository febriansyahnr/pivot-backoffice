package callbackWorker

import (
	"context"
	"encoding/base64"
	"maps"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"

	modelSdk "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) WriteCallbackLog(ctx context.Context, task *conductor.Task) (*conductor.TaskResult, error) {
	ctx, span := otelTracer.Start(ctx, "port/worker/callback/WriteCallbackLog")
	defer span.End()

	request := model.WorkflowWriteLogRequest{}
	if err := h.jsonBinder.Bind(&request, task.InputData); err != nil {
		h.logger.Error(ctx, "Failed while binding input data to request model", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}
	request.WorkflowId = task.WorkflowInstanceId
	request.RawPayload, _ = base64.StdEncoding.DecodeString(request.Payload)
	if request.Iteration > 0 {
		request.RetryCount = request.Iteration - 1
	}

	response, err := h.service.WriteCallbackLogFromWorkflowTask(ctx, request)
	if err != nil {
		h.logger.Error(ctx, "Failed while write callback log from workflow task", logger.Error(err))
		return conductor.NewTaskResultWithError(task, err), nil
	}

	taskResult := conductor.NewTaskResult(task)

	taskResult.Status = modelSdk.CompletedTask
	taskResult.OutputData, _ = modelSdk.ConvertToMap(response)

	return taskResult, nil
}

func (h *handler) WriteCallbackMetric(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error) {
	ctx, span := otelTracer.Start(ctx, "port/worker/callback/WriteCallbackMetric")
	defer span.End()

	request := model.WorkflowRecordMetricRequest{}
	if err = h.jsonBinder.Bind(&request, task.InputData); err != nil {
		h.logger.Error(ctx, "Failed while binding input data to request model", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}

	h.recordMetric(ctx, request)

	return noop(ctx, task)
}

func (h *handler) recordMetric(ctx context.Context, request model.WorkflowRecordMetricRequest) {

	// Metric attributes
	newAttributes := func(customs map[string]any) map[string]any {
		attrs := map[string]any{
			"merchantId": request.MerchantId,
			"eventName":  request.EventName,
			"statusCode": request.StatusCode,
		}

		maps.Copy(attrs, customs)

		return attrs
	}

	// Set the number of retries if any
	if (request.Iteration-1) > 0 && request.RetryCount == 0 {
		request.RetryCount = 1
	}

	// Metrics data
	metrics := make([]monitoring.CustomMetric, 0, 2)

	// Metric to record the number of callbacks sent
	if (request.Iteration + request.RetryCount) == 0 {
		metrics = append(metrics, monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentMerchantCallback,
			MetricName:           constant.MetricNameMerchantCallbackCount,
			Attributes:           newAttributes(map[string]any{"errorDetail": request.ErrorDetail}),
		})
	}

	// Metrics to record the duration of the callback sending process
	if request.DurationMs > 0 {
		metrics = append(metrics, monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeHistogram,
			MetricValue:          request.DurationMs,
			ComponentName:        constant.ComponentMerchantCallback,
			MetricName:           constant.MetricNameMerchantCallbackDuration,
			Attributes:           newAttributes(nil),
		})
	}

	// Metric to record the number of callback attempts
	if request.RetryCount > 0 {
		metrics = append(metrics, monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          request.RetryCount,
			ComponentName:        constant.ComponentMerchantCallback,
			MetricName:           constant.MetricNameMerchantCallbackRetryCount,
			Attributes:           newAttributes(nil),
		})
	}

	for _, metric := range metrics {
		if err := customMetric.RecordCustomMetric(ctx, &metric); err != nil {
			h.logger.Warn(ctx, "Failed while sending metrics on workflow write metric", logger.Error(err))
		}
	}
}
