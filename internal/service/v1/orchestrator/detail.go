package orchestrator_service

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *OrchestratorService) GetDetailById(ctx context.Context, merchantId, id string) (*orchestratorModel.TransactionHistoryDetailResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GetDetailById")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	resp, err := s.accountTransactionRepo.GetDetailById(ctx, merchantId, id)
	if err != nil {
		s.logger.Error(ctx, "get detail transaction with id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if resp == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	return resp, nil
}
