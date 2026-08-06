package industry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/country"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAll(t *testing.T) {
	service := serviceMocks.NewICountryService(t)

	router := chi.NewRouter()
	router.Get("/countries", New(service, validatorExt.New()).GetAll)

	countries := []*countryModel.Country{
		{Code: "ID", Name: "Indonesia", NameID: "Indonesia"},
		{Code: "SG", Name: "Singapore", NameID: "Singapura"},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Service error",
			queryParams: "",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Get all countries without filter",
			queryParams: "",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(countries, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"code":"ID","name":"Indonesia","nameId":"Indonesia","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"},{"code":"SG","name":"Singapore","nameId":"Singapura","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:        "SUCCESS: Get countries with English name filter",
			queryParams: "?name=Indonesia",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything).Once().Return([]*countryModel.Country{countries[0]}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"code":"ID","name":"Indonesia","nameId":"Indonesia","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:        "SUCCESS: Get countries with Indonesian name filter",
			queryParams: "?name=Indonesia&lang=id",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything).Once().Return([]*countryModel.Country{countries[0]}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"code":"ID","name":"Indonesia","nameId":"Indonesia","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:        "SUCCESS: Get countries with Indonesian name filter (uppercase language)",
			queryParams: "?name=Indonesia&lang=ID",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything).Once().Return([]*countryModel.Country{countries[0]}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"code":"ID","name":"Indonesia","nameId":"Indonesia","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:        "SUCCESS: Get countries with English name filter (non-ID language)",
			queryParams: "?name=Singapore&lang=en",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything).Once().Return([]*countryModel.Country{countries[1]}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"code":"SG","name":"Singapore","nameId":"Singapura","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
		{
			name:        "SUCCESS: Empty result",
			queryParams: "?name=NonExistent",
			setupMock: func() {
				service.On(
					"GetAll", c.ValueCtxMockType(), mock.Anything).Once().Return([]*countryModel.Country{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/countries"+test.queryParams, nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
		})
	}
}
