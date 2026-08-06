package location_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/location"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	service := serviceMocks.NewIAddrLocationService(t)

	router := chi.NewRouter()
	router.Get("/locations/{name}", New(validatorExt.New(), service).Get)

	tests := []struct {
		name           string
		locName        string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid location name",
			locName:        "X",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Name":"Key: 'LocationReq.Name' Error:Field validation for 'Name' failed on the 'oneof' tag"}}`,
		},
		{
			name:           "ERROR:Invalid province id",
			locName:        c.CityName,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"ProvinceId":"Key: 'LocationReq.ProvinceId' Error:Field validation for 'ProvinceId' failed on the 'required_if' tag"}}`,
		},
		{
			name:           "ERROR:Invalid city id",
			locName:        c.DistrictName,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"CityId":"Key: 'LocationReq.CityId' Error:Field validation for 'CityId' failed on the 'required_if' tag"}}`,
		},
		{
			name:    "ERROR:Some error",
			locName: c.ProvinceName,
			setupMock: func() {
				service.On(
					"Get", c.ValueCtxMockType(), mock.AnythingOfType("*location.LocationReq"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:    "SUCCESS",
			locName: c.ProvinceName,
			setupMock: func() {
				service.On(
					"Get", c.ValueCtxMockType(), mock.AnythingOfType("*location.LocationReq"),
				).Return(&location.LocationResp{ProvinceList: []location.Province{}}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"provinceList":[]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/locations/"+test.locName, nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
