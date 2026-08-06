package industry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	mockService := serviceMocks.NewIIndustryService(t)
	validate := validatorExt.New()

	router := chi.NewRouter()
	router.Delete("/industry/{id}", NewController(mockService, validate).Delete)

	tests := []struct {
		name           string
		industryID     string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
		isJSON         bool
	}{
		{
			name:       "SUCCESS: Delete industry",
			industryID: "test-uuid-123",
			setupMock: func() {
				mockService.On(
					"DeleteIndustry", mock.Anything, "test-uuid-123",
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":null}`,
			isJSON:         true,
		},
		{
			name:           "ERROR: Missing industry ID",
			industryID:     "",
			setupMock:      func() {},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   "404 page not found\n",
			isJSON:         false,
		},
		{
			name:       "ERROR: Industry not found",
			industryID: "non-existent-uuid",
			setupMock: func() {
				mockService.On(
					"DeleteIndustry", mock.Anything, "non-existent-uuid",
				).Once().Return(pkgErrs.New(response.HttpErrRequest, c.ErrIndustryNotFound))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"industry not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:       "ERROR: Industry in use by merchants",
			industryID: "test-uuid-123",
			setupMock: func() {
				mockService.On(
					"DeleteIndustry", mock.Anything, "test-uuid-123",
				).Once().Return(pkgErrs.New(response.HttpErrRequest, c.ErrIndustryInUse))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"cannot delete industry that is in use by merchants","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
		{
			name:       "ERROR: Service error",
			industryID: "test-uuid-123",
			setupMock: func() {
				mockService.On(
					"DeleteIndustry", mock.Anything, "test-uuid-123",
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
			isJSON:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			url := "/industry/" + test.industryID
			req := httptest.NewRequest(http.MethodDelete, url, nil)

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.industryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if test.isJSON {
				assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
			} else {
				assert.Equal(t, test.wantRespBody, rec.Body.String())
			}
		})
	}
}
