package country_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/country"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAll(t *testing.T) {
	log := loggerMocks.NewILogger(t)
	repo := repoMocks.NewICountryRepository(t)

	service := New(repo, log)

	countries := []*countryModel.Country{
		{Code: "ID", Name: "Indonesia", NameID: "Indonesia"},
		{Code: "SG", Name: "Singapore", NameID: "Singapura"},
	}

	tests := []struct {
		name       string
		filter     *countryModel.SearchFilterRequest
		setupMock  func()
		wantErr    error
		wantResult []*countryModel.Country
	}{
		{
			name:   "ERROR: Repository error",
			filter: &countryModel.SearchFilterRequest{Name: "Indonesia"},
			setupMock: func() {
				repo.On("GetAll", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrGetAllCountries),
		},
		{
			name:   "SUCCESS: Get all countries with filter",
			filter: &countryModel.SearchFilterRequest{Name: "Indonesia"},
			setupMock: func() {
				repo.On("GetAll", mock.Anything, mock.Anything).Once().Return(countries, nil)
			},
			wantResult: countries,
		},
		{
			name:   "SUCCESS: Get all countries without filter",
			filter: nil,
			setupMock: func() {
				repo.On("GetAll", mock.Anything, mock.Anything).Once().Return(countries, nil)
			},
			wantResult: countries,
		},
		{
			name:   "SUCCESS: Empty result",
			filter: &countryModel.SearchFilterRequest{Name: "NonExistent"},
			setupMock: func() {
				repo.On("GetAll", mock.Anything, mock.Anything).Once().Return([]*countryModel.Country{}, nil)
			},
			wantResult: []*countryModel.Country{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetAll(context.Background(), test.filter)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestFindByCode(t *testing.T) {
	log := loggerMocks.NewILogger(t)
	repo := repoMocks.NewICountryRepository(t)

	service := New(repo, log)

	country := &countryModel.Country{
		Code:   "ID",
		Name:   "Indonesia",
		NameID: "Indonesia",
	}

	tests := []struct {
		name       string
		code       string
		setupMock  func()
		wantErr    error
		wantResult *countryModel.Country
	}{
		{
			name: "ERROR: Repository error",
			code: "ID",
			setupMock: func() {
				repo.On("FindByCode", mock.Anything, "ID").Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrGetCountryByCode),
		},
		{
			name: "SUCCESS: Country found",
			code: "ID",
			setupMock: func() {
				repo.On("FindByCode", mock.Anything, "ID").Once().Return(country, nil)
			},
			wantResult: country,
		},
		{
			name: "SUCCESS: Different country code",
			code: "SG",
			setupMock: func() {
				sgCountry := &countryModel.Country{
					Code:   "SG",
					Name:   "Singapore",
					NameID: "Singapura",
				}
				repo.On("FindByCode", mock.Anything, "SG").Once().Return(sgCountry, nil)
			},
			wantResult: &countryModel.Country{
				Code:   "SG",
				Name:   "Singapore",
				NameID: "Singapura",
			},
		},
		{
			name: "SUCCESS: Empty string code",
			code: "",
			setupMock: func() {
				repo.On("FindByCode", mock.Anything, "").Once().Return(nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.FindByCode(context.Background(), test.code)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
