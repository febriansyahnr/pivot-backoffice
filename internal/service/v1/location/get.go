package location

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *locationService) Get(ctx context.Context, req *location.LocationReq) (res *location.LocationResp, err error) {
	res = &location.LocationResp{}

	isFound := true
	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	switch req.Name {
	case constant.ProvinceName:
		res.ProvinceList, err = s.repo.GetProvinces(ctx)

	case constant.CityName:
		provinceId, _ := strconv.Atoi(req.ProvinceId)
		res.CityList, err = s.repo.GetCitiesByProvinceId(ctx, util.MustIntToUint16(provinceId))

		isFound = !reflect.ValueOf(res.CityList).IsNil()

	case constant.DistrictName:
		cityId, _ := strconv.Atoi(req.CityId)
		res.DistrictList, err = s.repo.GetDistrictsByCityId(ctx, util.MustIntToUint16(cityId))

		isFound = !reflect.ValueOf(res.DistrictList).IsNil()
	}

	if err != nil {
		s.logger.Error(ctx, "get data location "+req.Name, logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if !isFound {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	return
}
