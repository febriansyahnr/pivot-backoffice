package industry

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// CreateIndustry creates a new industry record
func (s *IndustryService) CreateIndustry(ctx context.Context, req industryModel.CreateIndustryRequest) (*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.CreateIndustry")
	defer span.End()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, err)
	}

	// Check for duplicate parent-child combination
	existing, err := s.repo.GetByParentChildIndustry(ctx, req.ParentIndustry, req.ChildIndustry)
	if err != nil {
		s.logger.Error(ctx, "failed to check duplicate industry", logger.Error(err), logger.String("parentIndustry", req.ParentIndustry), logger.String("childIndustry", req.ChildIndustry))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrCreateIndustry)
	}
	if existing != nil {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateIndustry)
	}

	// Create the industry record using model factory
	industry := industryModel.NewIndustry(&req)

	if err := s.repo.Create(ctx, industry); err != nil {
		s.logger.Error(ctx, "failed to create industry", logger.Error(err), logger.Any("industry", industry))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrCreateIndustry)
	}

	return industry, nil
}
