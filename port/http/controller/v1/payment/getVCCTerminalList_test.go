package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVCCTerminalList(t *testing.T) {
	const (
		merchantID = "550e8400-e29b-41d4-a716-446655440000"
		userID     = "550e8400-e29b-41d4-a716-446655440001"
		chargeID   = "550e8400-e29b-41d4-a716-446655440002"
	)

	userInfo := &userModel.UserTokenClaims{
		UUID:       userID,
		MerchantId: merchantID,
	}
	paymentService := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validatorExt.New(), nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/vcc-terminal/charges", handler.GetVCCTerminalList)

	chargeDate := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		userInfo         *userModel.UserTokenClaims
		queryParams      string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR: User not found in context", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid chargeStartDate format", // NOSONAR
			userInfo:         userInfo,
			queryParams:      "?chargeStartDate=invalid-date",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid chargeStartDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid chargeEndDate format", // NOSONAR
			userInfo:         userInfo,
			queryParams:      "?chargeEndDate=invalid-date",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid chargeEndDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid page format", // NOSONAR
			userInfo:         userInfo,
			queryParams:      "?page=invalid",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid page format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid limit format", // NOSONAR
			userInfo:         userInfo,
			queryParams:      "?limit=invalid",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid limit format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service returns error", // NOSONAR
			userInfo:    userInfo,
			queryParams: "?status=SUCCESS",
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Get VCC Terminal list with status filter", // NOSONAR
			userInfo:    userInfo,
			queryParams: "?status=SUCCESS&page=1&limit=10&sortBy=chargeDate&sort=DESC",
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetVCCTerminalListFilterRequest) bool {
					return req.MerchantID == merchantID &&
						req.Status == "SUCCESS" &&
						req.Page == 1 &&
						req.Limit == 10 &&
						req.SortBy == "chargeDate" &&
						req.Sort == "DESC"
				})).Once().Return(&commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{
						{
							BulkID:     "bulk-001",
							ChargeID:   chargeID,
							ChargeDate: chargeDate,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(150000),
								Currency: "IDR",
							},
							Status:      "SUCCESS",
							TravelAgent: "GARUDA",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[{"bulkId":"bulk-001","chargeId":"550e8400-e29b-41d4-a716-446655440002","chargeDate":"2024-01-01T10:00:00Z","amount":{"value":"150000","currency":"IDR"},"status":"SUCCESS","travelAgent":"GARUDA","referenceId":"","bookingId":""}],"pagination":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}`,
		},
		{
			name:        "SUCCESS: Get VCC Terminal list with chargeId filter", // NOSONAR
			userInfo:    userInfo,
			queryParams: "?chargeId=" + chargeID,
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetVCCTerminalListFilterRequest) bool {
					return req.MerchantID == merchantID && req.ChargeID == chargeID
				})).Once().Return(&commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{
						{
							BulkID:     "bulk-002",
							ChargeID:   chargeID,
							ChargeDate: chargeDate,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(200000),
								Currency: "IDR",
							},
							Status:      "PENDING",
							TravelAgent: "GARUDA",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    1,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[{"bulkId":"bulk-002","chargeId":"550e8400-e29b-41d4-a716-446655440002","chargeDate":"2024-01-01T10:00:00Z","amount":{"value":"200000","currency":"IDR"},"status":"PENDING","travelAgent":"GARUDA","referenceId":"","bookingId":""}],"pagination":{"page":1,"perPage":1,"totalItems":1,"totalPages":1}}`,
		},
		{
			name:        "SUCCESS: Get VCC Terminal list with date range filter", // NOSONAR
			userInfo:    userInfo,
			queryParams: "?chargeStartDate=2024-01-01T00:00:00Z&chargeEndDate=2024-01-31T23:59:59Z",
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetVCCTerminalListFilterRequest) bool {
					startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
					endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
					return req.MerchantID == merchantID &&
						req.ChargeStartDate.Equal(startDate) &&
						req.ChargeEndDate.Equal(endDate)
				})).Once().Return(&commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    1,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":1,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:        "SUCCESS: Get VCC Terminal list with perPage parameter", // NOSONAR
			userInfo:    userInfo,
			queryParams: "?perPage=20",
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetVCCTerminalListFilterRequest) bool {
					return req.MerchantID == merchantID && req.Limit == 20
				})).Once().Return(&commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{
						{
							BulkID:     "bulk-003",
							ChargeID:   chargeID,
							ChargeDate: chargeDate,
							Amount: paymentModel.Amount{
								Value:    decimal.NewFromInt(100000),
								Currency: "IDR",
							},
							Status:      "SUCCESS",
							TravelAgent: "AIRASIA",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[{"bulkId":"bulk-003","chargeId":"550e8400-e29b-41d4-a716-446655440002","chargeDate":"2024-01-01T10:00:00Z","amount":{"value":"100000","currency":"IDR"},"status":"SUCCESS","travelAgent":"AIRASIA","referenceId":"","bookingId":""}],"pagination":{"page":1,"perPage":20,"totalItems":1,"totalPages":1}}`,
		},
		{
			name:        "SUCCESS: Get VCC Terminal list with no filters", // NOSONAR
			userInfo:    userInfo,
			queryParams: "",
			setupMock: func() {
				paymentService.On("GetVCCTerminalList", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetVCCTerminalListFilterRequest) bool {
					return req.MerchantID == merchantID &&
						req.Status == "" &&
						req.ChargeID == "" &&
						req.ChargeStartDate.IsZero() &&
						req.ChargeEndDate.IsZero()
				})).Once().Return(&commonModel.PaginationResponse{
					Data: []paymentModel.VccTerminalItem{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    1,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":1,"totalItems":0,"totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/vcc-terminal/charges"+test.queryParams, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userInfo != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userInfo))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}
