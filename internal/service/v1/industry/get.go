package industry

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *IndustryService) GetAllIndustries(ctx context.Context, request *industryModel.SearchIndustryRequest) ([]*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.GetAllIndustries")
	defer span.End()

	data, err := s.repo.GetAllIndustries(ctx, request)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrGetAllCountries)
	}
	return data, nil
}

func (s *IndustryService) GetUniqueParentIndustries(ctx context.Context) ([]string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.GetUniqueParentIndustries")
	defer span.End()

	return s.repo.GetUniqueParentIndustries(ctx)
}

func (s *IndustryService) GetChildIndustries(ctx context.Context, parentIndustry string) ([]string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.GetChildIndustries")
	defer span.End()

	return s.repo.GetChildIndustries(ctx, parentIndustry)
}

func (s *IndustryService) GetMCCForIndustry(ctx context.Context, parentIndustry, childIndustry string) (string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.GetMCCForIndustry")
	defer span.End()

	return s.repo.GetMCCForIndustry(ctx, parentIndustry, childIndustry)
}

func (s *IndustryService) GetIndustryByID(ctx context.Context, id string) (*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.GetIndustryByID")
	defer span.End()

	industry, err := s.repo.GetIndustryByID(ctx, id)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrGetIndustryByID)
	}

	return industry, nil
}
