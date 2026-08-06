package fraudruleservice

import (
	"context"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *FraudRuleService) Create(ctx context.Context, request *fraudrulesmodel.CreateFraudRuleRequest) (*fraudrulesmodel.FraudRulesResponse, error) {
	ctx, segment := tracer.Start(ctx, "internal/service/v1/fraudRule/Create")
	defer segment.End()

	fraudRule, err := fraudrulesmodel.New(request)
	if err != nil {
		return nil, errors.New(response.HttpErrRequest, err)
	}

	err = s.fraudRulesRepository.Create(ctx, fraudRule)
	if err != nil {
		return nil, errors.New(response.HttpErrInternal, err)
	}

	return fraudRule.ToResponse(), nil
}
