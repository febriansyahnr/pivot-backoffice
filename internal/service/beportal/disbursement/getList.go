package disbursementService

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *DisbursementService) GetList(
	ctx context.Context,
	filter *disbursementModel.GetDisbursementFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetList")
	defer segment.End()

	err := filter.Validate()
	if err != nil {
		return nil, errPkg.New(response.HttpErrRequest, err)
	}

	list, err := s.disbursementRepo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
