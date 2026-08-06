package country

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
)

func (s *countryService) GetAll(ctx context.Context, filter *countryModel.SearchFilterRequest) ([]*countryModel.Country, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/country/GetAll")
	defer segment.End()

	countries, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetAllCountries)
	}

	return countries, nil
}

func (s *countryService) FindByCode(ctx context.Context, code string) (*countryModel.Country, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/country/FindByCode")
	defer segment.End()

	country, err := s.repo.FindByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetCountryByCode)
	}

	return country, nil
}
