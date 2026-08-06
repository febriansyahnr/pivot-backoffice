package callbackService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *CallbackService) GetCallbackLogList(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetCallbackLogList")
	defer segment.End()

	result, err := s.callbackRepo.FindMerchantCallbackHistory(ctx, filter, page, perPage)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return result, nil
}
