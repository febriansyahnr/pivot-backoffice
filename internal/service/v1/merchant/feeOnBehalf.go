package merchant

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) CreateFeeConfigOnBehalf(ctx context.Context, request *merchant.CreateFeeConfigOnBehalfRequest) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CreateFeeConfigOnBehalf")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	valid, err := s.repo.ValidateCreateFeeConfigOnBehalf(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "Failed when validation create fee on-behalf", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if !valid {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("your request is invalid, please check your data again"))
	}

	data := request.ToOnBehalfFeeConfig()

	if err := s.repo.CreateFeeConfigOnBehalf(ctx, data); err != nil {
		s.logger.Error(ctx, "Failed when create fee config on-behalf", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	return data.Id, nil
}

func (s *MerchantService) GetFeeConfigOnBehalf(ctx context.Context, request *merchant.GetFeeConfigOnBehalfRequest) ([]merchant.FeeConfigOnBehalfResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetFeeConfigOnBehalf")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	configs, err := s.repo.GetFeeConfigOnBehalf(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "Failed when get fee config on-behalf", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	return configs, nil
}

func (s *MerchantService) UpdateFeeConfigOnBehalf(ctx context.Context, id string, request *merchant.UpdateFeeConfigOnBehalfRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateFeeConfigOnBehalf")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	if err := s.repo.UpdateFeeConfigOnBehalf(ctx, id, request); errors.Is(err, constant.ErrNoRowsAffected) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)

	} else if err != nil {
		s.logger.Error(ctx, "Failed when update fee config on-behalf", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	return nil
}
