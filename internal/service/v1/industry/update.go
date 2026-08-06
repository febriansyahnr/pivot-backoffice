package industry

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// UpdateIndustry updates an existing industry record
func (s *IndustryService) UpdateIndustry(ctx context.Context, req industryModel.UpdateIndustryRequest) (*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.UpdateIndustry")
	defer span.End()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, err)
	}

	// Get existing industry
	existing, err := s.repo.GetIndustryByID(ctx, req.UUID)
	if err != nil {
		s.logger.Error(ctx, "failed to get industry by id", logger.Error(err), logger.String("uuid", req.UUID))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry)
	}
	if existing == nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryNotFound)
	}

	// Check for duplicate parent-child combination if updating those fields
	if req.ParentIndustry != nil || req.ChildIndustry != nil {
		parent := existing.ParentIndustry
		child := existing.ChildIndustry
		if req.ParentIndustry != nil {
			parent = *req.ParentIndustry
		}
		if req.ChildIndustry != nil {
			child = *req.ChildIndustry
		}

		duplicate, err := s.repo.GetByParentChildIndustry(ctx, parent, child)
		if err != nil {
			s.logger.Error(ctx, "failed to check duplicate industry", logger.Error(err), logger.String("parentIndustry", parent), logger.String("childIndustry", child))
			return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry)
		}
		if duplicate != nil && duplicate.UUID != req.UUID {
			return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateIndustry)
		}
	}

	// Apply update using model method
	updated := existing.ApplyUpdate(&req)

	if err := s.repo.Update(ctx, updated); err != nil {
		s.logger.Error(ctx, "failed to update industry", logger.Error(err), logger.Any("industry", updated))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry)
	}

	return updated, nil
}