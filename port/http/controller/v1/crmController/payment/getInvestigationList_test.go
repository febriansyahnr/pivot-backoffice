package v1CrmPaymentController_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetInvestigationList(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	tests := []struct {
		name           string
		queryParams    string
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "SUCCESS: Get investigation list with default params",
			queryParams: "",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":1, "perPage":20, "totalItems":0, "totalPages":0}}`,
		},
		{
			name:        "SUCCESS: Get investigation list with status filter",
			queryParams: "investigationStatus=INVESTIGATION_IN_PROCESS",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":1, "perPage":20, "totalItems":0, "totalPages":0}}`,
		},
		{
			name:        "SUCCESS: Get investigation list with pagination",
			queryParams: "page=2&limit=50",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{Page: 2, PerPage: 50, TotalItems: 100, TotalPages: 2},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":2, "perPage":50, "totalItems":100, "totalPages":2}}`,
		},
		{
			name:        "SUCCESS: Get investigation list with date range",
			queryParams: "fromDate=2026-01-01T00:00:00Z&toDate=2026-01-31T23:59:59Z",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":1, "perPage":20, "totalItems":0, "totalPages":0}}`,
		},
		{
			name:        "SUCCESS: Get investigation list with sorting",
			queryParams: "sortBy=updatedAt&sort=ASC",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message":"OK", "pagination":{"page":1, "perPage":20, "totalItems":0, "totalPages":0}}`,
		},
		{
			name:           "ERROR: Invalid page format",
			queryParams:    "page=abc",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid page format. Use number format instead"}`,
		},
		{
			name:           "ERROR: Invalid limit format",
			queryParams:    "limit=abc",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid limit format. Use number format instead"}`,
		},
		{
			name:           "ERROR: Invalid fromDate format",
			queryParams:    "fromDate=invalid-date",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid fromDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}`,
		},
		{
			name:           "ERROR: Invalid toDate format",
			queryParams:    "toDate=invalid-date",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid toDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "",
			mockService: func() {
				svc.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			rec := httptest.NewRecorder()
			url := "/payments/investigations"
			if test.queryParams != "" {
				url = fmt.Sprintf("/payments/investigations?%s", test.queryParams)
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			router := chi.NewRouter()
			router.Get("/payments/investigations", h.GetInvestigationList)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
