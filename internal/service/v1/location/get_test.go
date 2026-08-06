package location_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/location"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	provinces = []location.Province{
		{
			Id:        15,
			Name:      "JAWA TIMUR",
			CreatedAt: time.Now().UTC(),
		},
	}
	cities = []location.City{
		{
			Id:         259,
			ProvinceId: 15,
			Name:       "KOTA MALANG",
			CreatedAt:  time.Now().UTC(),
		},
	}
	districts = []location.District{
		{
			Id:        3887,
			CityId:    259,
			Name:      "BLIMBING",
			CreatedAt: time.Now().UTC(),
		},
	}
)

func TestGet(t *testing.T) {
	repo := repoMocks.NewIAddrLocationRepository(t)

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	service := New(logger, repo)

	traceId := uuid.NewString()

	tests := []struct {
		name       string
		locName    string
		setupMock  func()
		wantErr    string
		wantResult *location.LocationResp
	}{
		{
			name:    "ERROR:Some error",
			locName: c.ProvinceName,
			setupMock: func() {
				repo.On("GetProvinces", c.ValueCtxMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name:    "ERROR:Data not found",
			locName: c.CityName,
			setupMock: func() {
				repo.On(
					"GetCitiesByProvinceId", c.ValueCtxMockType(), c.Uint16MockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name:    "SUCCESS:Provinces",
			locName: c.ProvinceName,
			setupMock: func() {
				repo.On("GetProvinces", c.ValueCtxMockType()).Return(provinces, nil)
			},
			wantResult: &location.LocationResp{
				ProvinceList: provinces,
			},
		},
		{
			name:    "SUCCESS:Cities",
			locName: c.CityName,
			setupMock: func() {
				repo.On("GetCitiesByProvinceId", c.ValueCtxMockType(), c.Uint16MockType()).Return(cities, nil)
			},
			wantResult: &location.LocationResp{
				CityList: cities,
			},
		},
		{
			name:    "SUCCESS:Districts",
			locName: c.DistrictName,
			setupMock: func() {
				repo.On("GetDistrictsByCityId", c.ValueCtxMockType(), c.Uint16MockType()).Return(districts, nil)
			},
			wantResult: &location.LocationResp{
				DistrictList: districts,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
			if resp, err := service.Get(ctx, &location.LocationReq{Name: test.locName}); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
