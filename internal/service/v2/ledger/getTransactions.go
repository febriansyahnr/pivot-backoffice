package ledgerService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *LedgerService) GetLedgerTransactions(ctx context.Context, request *ledger_model.GetLedgerTransactionRequest, pagination *commonModel.Meta) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledgerService/GetLedgerTransactions")
	defer segment.End()

	if err := request.AdjustDateTime(); err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	request.Status = constant.StatusSuccess

	ledgerData, total, err := s.repo.GetLedgerRecords(ctx, request, pagination)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetLedgerRecords)
	}
	return &commonModel.PaginationResponse{
		Data: ledgerData,
		Meta: *commonModel.NewMeta(pagination.Page, pagination.PerPage, int64(total)),
	}, nil

}
