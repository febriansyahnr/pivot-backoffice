package v1CrmPaymentController_test

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"
)

func TestGetList(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validRequest := paymentModel.GetListFilterRequest{
		Page:    1,
		PerPage: 10,
		Sort:    "ASC",
		SortBy:  "createdAt",
	}

	tests := []struct {
		name           string
		queryParams    string
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid page format",
			queryParams:    "page=abc",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid page format. Use number format instead"}`,
		},
		{
			name:           "ERROR: Invalid startDate format",
			queryParams:    "startDate=invalid-date",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "page=1&perPage=10",
			mockService: func() {
				svc.On("GetListForInternalDashboard", mock.Anything, mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			queryParams: "page=1&perPage=10",
			mockService: func() {
				svc.On("GetListForInternalDashboard", mock.Anything, &validRequest).
					Once().Return(&commonModel.PaginationResponse{Data: []interface{}{}, Meta: commonModel.Meta{}}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/payments?%s", test.queryParams), nil)

			router := chi.NewRouter()
			router.Get("/payments", h.GetList)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
