package callbackWorker

import (
	"context"
	"encoding/base64"
	"fmt"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	modelSdk "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (h *handler) Preparation(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error) {
	ctx, span := otelTracer.Start(ctx, "port/worker/callback/Preparation")
	defer func() {
		if r := recover(); r != nil {
			result = conductor.NewTaskResultWithNonRetryableError(task, fmt.Errorf("panic recovery: %v", r))
		}
		span.End()
	}()

	inputData := model.WorkflowMerchantCallbackRequest{}
	if err = h.jsonBinder.Bind(&inputData, task.InputData); err != nil {
		h.logger.Error(ctx, "Failed while binding input data to request model", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}

	protoBytes, err := base64.StdEncoding.DecodeString(inputData.Payload)
	if err != nil {
		h.logger.Error(ctx, "Failed while decode base64 payload to bytes", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}

	message := &pb.ProcessCallbackRequest{}
	if err = proto.Unmarshal(protoBytes, message); err != nil {
		h.logger.Error(ctx, "Failed while decompiling proto message", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}

	request := model.ProcessCallbackRequest{
		Name:       message.Name,
		Event:      message.Event,
		MerchantID: util.ParseUUID(message.MerchantId),
		IsSnap:     message.IsSnap,
	}
	if err = request.Bind(message.Request); err != nil {
		h.logger.Error(ctx, "Failed while binding message request to callback request", logger.Error(err))
		return conductor.NewTaskResultWithNonRetryableError(task, err), nil
	}

	callback, err := h.service.FindCallbackByMerchantIdAndCallbackName(ctx, request.MerchantID, request.Name)
	if err != nil {
		h.logger.Error(ctx, "Failed while find merchant callback details", logger.Error(err))
		return conductor.NewTaskResultWithError(task, err), nil

	} else if callback == nil {
		callback = &model.Callback{}
	}

	response := request.ToWorkflowPreparationResponse(callback.UUID.String(), callback.URL)

	h.logger.Info(
		ctx, "Data preparation for merchant callback delivery",
		logger.String("merchantId", response.MerchantId), logger.String("referenceId", util.ValueOfPtr(response.ReferenceId)), logger.String("workflowId", task.WorkflowInstanceId),
	)

	taskResult := conductor.NewTaskResult(task)

	taskResult.Status = modelSdk.CompletedTask
	taskResult.OutputData, _ = modelSdk.ConvertToMap(response)

	return taskResult, nil
}
