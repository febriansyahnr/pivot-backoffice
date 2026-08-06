package fraudruleservice

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *FraudRuleService) Detail(ctx context.Context, uuid string) (*fraudrulesmodel.FraudRules, error) {
	ctx, segment := tracer.Start(ctx, "internal/service/v1/fraudRule/Detail")
	defer segment.End()

	fraudRule, err := s.fraudRulesRepository.GetByID(ctx, uuid)
	if err != nil {
		return nil, errors.New(response.HttpErrInternal, constant.ErrGetFraudRuleDetail)
	}
	if fraudRule == nil {
		return nil, errors.New(response.HttpErrRequest, constant.ErrFraudRulesNotFound)
	}

	return fraudRule, nil
}
