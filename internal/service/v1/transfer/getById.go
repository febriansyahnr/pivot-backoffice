package transferService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *TransferService) GetById(ctx context.Context, id, merchantId string) (*transfer.Transfer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/GetById")
	defer segment.End()

	transfer, err := s.repo.GetByID(ctx, id, merchantId)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetTransferById)
	}
	if transfer == nil {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrTransferNotFound)
	}

	return transfer, nil
}

// GetTransferTransaction retrieves the details of a transfer transaction by its ID.
// If the transfer transaction is not found, it returns a not found error.
// If there is an internal error during the retrieval process, it returns an internal error.
func (s *TransferService) GetTransferTransaction(ctx context.Context, req transfer.GetTransferTransactionRequest) (*transfer.TransferTransactionDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/GetTransferTransaction")
	defer segment.End()

	transfer, err := s.repo.GetTransferTransaction(ctx, req)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, err)
	}

	if transfer == nil {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrTransferNotFound)
	}

	return transfer, nil
}
