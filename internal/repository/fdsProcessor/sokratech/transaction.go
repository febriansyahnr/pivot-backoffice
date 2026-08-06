package sokratech

import (
	"context"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
)

func (r *repository) AssessPayoutTransaction(ctx context.Context, request fdscommon.AssessPayoutTransactionRequest) (*fdscommon.TransactionAssessmentResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/fdsProcessor/sokratech/AssessPayoutTransaction")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.config.TimeoutSeconds)*time.Second)
	defer cancel()

	workflowRequest := &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
		WorkflowID:   r.config.Workflow.PayoutTransactionID,
		WorkflowName: constant.FDSWorkflowNamePayoutRules,
		Payload:      toPayoutWorkflowRequest(request),
	}

	result, err := r.workflowExecute(ctx, workflowRequest)
	if err != nil {
		return nil, err
	}

	return &fdscommon.TransactionAssessmentResponse{
		Result:              strings.ToUpper(result.Data.Result),
		RiskScore:           result.Data.Score,
		ExecutionID:         result.Data.SystemVariables.ExecutionID,
		WorkflowID:          result.Data.SystemVariables.WorkflowID,
		WorkflowIterationID: result.Data.SystemVariables.WorkflowIterationID,
		WorkflowVersion:     result.Data.SystemVariables.WorkflowVersion,
		ExecutionStartTime:  &result.Data.SystemVariables.ExecutionStartTime,
	}, nil
}
