package sokratech

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

type fdsProcessor struct {
	repo *repository
}

func (r *repository) NewFDSProcessor() port.IFdsProcessorRepository {
	return &fdsProcessor{repo: r}
}

func (p *fdsProcessor) Check(ctx context.Context, request *fdscommon.CheckRequest) (*fdscommon.CheckResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.repo.config.TimeoutSeconds)*time.Second)
	defer cancel()

	workflowRequest := &model.WorkflowExecuteRequest[model.PaymentWorkflowRequest]{
		WorkflowID:   p.repo.config.Workflow.PaymentTransactionID,
		WorkflowName: constant.FDSWorkflowNamePaymentRules,
		Payload:      toPaymentWorkflowRequest(request),
	}
	result, err := p.repo.workflowExecute(ctx, workflowRequest)
	if err != nil {
		if groupErr, _ := pkgErrs.ExtractError(err); groupErr == response.HttpErrInternal {
			return nil, err
		}
		return &fdscommon.CheckResponse{Success: false}, nil
	}
	return &fdscommon.CheckResponse{
		Success: true,
		Data: fdscommon.CheckData{
			ID:        result.Data.SystemVariables.ExecutionID,
			RiskScore: result.Data.Score,
			RiskGroup: result.Data.Result,
		},
	}, nil
}

func (p *fdsProcessor) Update(ctx context.Context, request *fdscommon.UpdateRequest) (*fdscommon.UpdateResponse, error) {
	if request == nil || request.FullContext == nil {
		return &fdscommon.UpdateResponse{Success: false}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.repo.config.TimeoutSeconds)*time.Second)
	defer cancel()

	payload := toPaymentWorkflowRequest(request.FullContext)
	if request.IsFraud != nil {
		payload.Transaction.IsFraud = request.IsFraud
		if request.Payment != nil {
			payload.Transaction.PaymentStatus = request.Payment.PaymentStatus
			payload.Transaction.ChargebackStatus = util.ValueOfPtr(request.Payment.ChargebackStatus)
			payload.Transaction.ChargebackNotes = util.ValueOfPtr(request.Note)
		}
	}

	workflowRequest := &model.WorkflowExecuteRequest[model.PaymentWorkflowRequest]{
		WorkflowID:   p.repo.config.Workflow.PaymentTransactionID,
		WorkflowName: constant.FDSWorkflowNamePaymentRules,
		Payload:      payload,
	}
	result, err := p.repo.workflowExecute(ctx, workflowRequest)
	if err != nil {
		if groupErr, _ := pkgErrs.ExtractError(err); groupErr == response.HttpErrInternal {
			return nil, err
		}
		return &fdscommon.UpdateResponse{Success: false}, nil
	}
	return &fdscommon.UpdateResponse{
		Success: true,
		Data: fdscommon.UpdateData{
			ID: result.Data.SystemVariables.ExecutionID,
		},
	}, nil
}
