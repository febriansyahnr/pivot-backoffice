package crmCreditcardController

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetTransactionList(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

	tests := []struct {
		name           string
		queryParams    string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Invalid page parameter",
			queryParams: "page=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid page number"}`,
		},
		{
			name:        "ERROR: Invalid perPage parameter",
			queryParams: "perPage=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid per page number"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "dateFrom=2023-01-01&dateTo=2023-12-31",
			modifierMock: func() {
				svc.On("GetTransactionList", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			queryParams: "page=1&perPage=10&dateFrom=2023-01-01&dateTo=2023-12-31",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetTransactionList", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"data":[], "meta":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}}`,
		},
		{
			name:        "SUCCESS: With charge date filters",
			queryParams: "page=1&perPage=10&chargeFrom=2023-01-01&chargeTo=2023-12-31",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetTransactionList", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"data":[], "meta":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}}`,
		},
		{
			name:        "SUCCESS: With refund date filters",
			queryParams: "page=1&perPage=10&refundFrom=2023-01-01&refundTo=2023-12-31",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetTransactionList", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"data":[], "meta":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}}`,
		},
		{
			name:        "SUCCESS: With all date filters combined",
			queryParams: "page=1&perPage=10&dateFrom=2023-01-01&dateTo=2023-12-31&chargeFrom=2023-02-01&chargeTo=2023-11-30&refundFrom=2023-03-01&refundTo=2023-10-31",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetTransactionList", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"data":[], "meta":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/crm/v1/creditcard/transaction/list?%s", test.queryParams), nil)

			router := chi.NewRouter()
			router.Get("/crm/v1/creditcard/transaction/list", New(&config.Config{}, &config.Secret{}, svc).GetTransactionList)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
