package fraudruleservice

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *FraudRuleService) Delete(ctx context.Context, uuid string) error {
	ctx, segment := tracer.Start(ctx, "internal/service/v1/fraudRule/Delete")
	defer segment.End()

	fraudRule, err := s.fraudRulesRepository.GetByID(ctx, uuid)
	if err != nil {
		return errors.New(response.HttpErrInternal, constant.ErrGetFraudRuleList)
	}
	if fraudRule == nil {
		return errors.New(response.HttpErrRequest, constant.ErrFraudRulesNotFound)
	}

	err = s.fraudRulesRepository.Delete(ctx, uuid)
	if err != nil {
		return errors.New(response.HttpErrInternal, constant.ErrDeleteFraudRule)
	}

	return nil
}
