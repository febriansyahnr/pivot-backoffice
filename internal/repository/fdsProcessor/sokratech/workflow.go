package sokratech

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *repository) workflowExecute(ctx context.Context, request WorkflowRequester) (result *model.WorkflowExecuteResponse, err error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/fdsProcessor/sokratech/workflowExecute", trace.WithAttributes(
		attribute.String("workflow.id", request.GetWorkflowID()),
		attribute.String("workflow.name", request.GetWorkflowName()),
		attribute.Int("status.code", 0),
	))
	defer span.End()

	var (
		start   = time.Now()
		uri     = r.config.BaseURL + "/public/workflows/" + request.GetWorkflowID() + "/execute"
		headers = map[string]string{
			constant.HeaderXAPIKey: r.secret.AccessSecret,
		}
		payload       = request.GetWorkflowPayload()
		responseBytes []byte
		statusCode    int
	)
	defer func() {
		duration := time.Since(start)
		span.SetAttributes(attribute.Int("status.code", statusCode))

		r.logger.Info(
			ctx, "Executing workflow with ID "+request.GetWorkflowID(),
			logger.Any("requestBody", payload),
			logger.Any("response", map[string]any{"body": string(responseBytes), "statusCode": statusCode}),
			logger.Any("duration", map[string]any{"human": duration.String(), "milliseconds": duration.Milliseconds()}),
		)
	}()

	responseBytes, statusCode, err = r.client.POST(ctx, uri, request.GetWorkflowPayload(), headers)
	if err != nil {
		r.logger.Error(ctx, "Failed to request workflow execution", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)

	} else if statusCode >= 500 {
		return nil, pkgErrs.New(response.HttpErrThirdParty, errors.New(string(responseBytes)))

	} else if statusCode >= 400 {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New(string(responseBytes)))
	}

	result = &model.WorkflowExecuteResponse{}
	if err = json.Unmarshal(responseBytes, result); err != nil {
		r.logger.Error(ctx, "Failed to unmarshal workflow response", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInvalidUnmarshalJSON)
	}

	return result, nil
}
