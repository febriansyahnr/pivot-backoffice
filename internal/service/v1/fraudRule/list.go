package fraudruleservice

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *FraudRuleService) List(ctx context.Context, req *fraudrulesmodel.FraudRulesQuery) (*commonModel.PaginationResponse, error) {
	ctx, segment := tracer.Start(ctx, "internal/service/v1/ipWhitelist/List")
	defer segment.End()

	list, total, err := s.fraudRulesRepository.List(ctx, req)
	if err != nil {
		return nil, errors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationList)
	}

	return &commonModel.PaginationResponse{
		Data: list,
		Meta: *commonModel.NewMeta(req.Page, req.PageSize, int64(total)),
	}, nil
}
