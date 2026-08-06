package callbackService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) GetCallbackLogDetail(ctx context.Context, id, merchantID string) (*callbackModel.CallbackLogWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetCallbackLogDetail")
	defer segment.End()

	res, err := s.callbackRepo.GetCallbackLogByID(ctx, id)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if res == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	if res.MerchantID != merchantID {
		s.logger.Error(ctx, constant.ErrMerchantIsNotMatch.Error(), logger.Error(constant.ErrMerchantIsNotMatch))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	return res, nil
}
