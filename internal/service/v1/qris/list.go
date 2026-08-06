package qris

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *qrisService) RegistrationList(ctx context.Context, merchantId string) ([]qris.RegistrationListResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/qris/RegistrationList")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	result, err := s.repository.RegistrationList(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "registration list by merchant id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	return result, nil
}
