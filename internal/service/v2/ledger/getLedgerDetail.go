package ledgerService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *LedgerService) GetLedgerDetail(ctx context.Context, referenceId string) ([]orchestrator_model.AccountTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledgerService/GetLedgerDetail")
	defer segment.End()

	ledgerDetails, err := s.repo.GetLedgerDetail(ctx, referenceId)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetLedgerDetail)
	}
	if ledgerDetails == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrLedgerDetailNotFound)
	}

	return ledgerDetails, nil

}
