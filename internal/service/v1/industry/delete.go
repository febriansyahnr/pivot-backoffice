package industry

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// DeleteIndustry soft deletes an industry record
func (s *IndustryService) DeleteIndustry(ctx context.Context, uuid string) error {
	ctx, span := otelTracer.Start(ctx, "IndustryService.DeleteIndustry")
	defer span.End()

	// Get existing industry
	existing, err := s.repo.GetIndustryByID(ctx, uuid)
	if err != nil {
		s.logger.Error(ctx, "failed to get industry by id for delete", logger.Error(err), logger.String("uuid", uuid))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry)
	}
	if existing == nil {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryNotFound)
	}

	// Check if industry is being used by merchants
	isUsed, err := s.repo.IsIndustryUsedByMerchants(ctx, existing.ParentIndustry, existing.ChildIndustry)
	if err != nil {
		s.logger.Error(ctx, "failed to check if industry is used by merchants", logger.Error(err), logger.String("uuid", uuid))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry)
	}
	if isUsed {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryInUse)
	}

	// Delete the industry
	if err := s.repo.Delete(ctx, uuid); err != nil {
		s.logger.Error(ctx, "failed to delete industry", logger.Error(err), logger.String("uuid", uuid))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry)
	}

	return nil
}