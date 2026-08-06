package fraudruleservice

import (
	"context"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *FraudRuleService) Update(ctx context.Context, request *fraudrulesmodel.UpdateFraudRuleRequest) (*fraudrulesmodel.FraudRulesResponse, error) {
	ctx, segment := tracer.Start(ctx, "internal/service/v1/fraudRule/Create")
	defer segment.End()

	// Get existing rule
	existing, err := s.fraudRulesRepository.GetByID(ctx, request.UUID)
	if err != nil {
		return nil, errors.New(response.HttpErrNotFound, err)
	}

	existing.Update(request)
	err = s.fraudRulesRepository.Update(ctx, existing)
	if err != nil {
		return nil, errors.New(response.HttpErrInternal, err)
	}

	return existing.ToResponse(), nil
}
