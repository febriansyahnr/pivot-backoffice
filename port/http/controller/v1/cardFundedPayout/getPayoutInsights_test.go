package cardFundedPayoutController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutInsights(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe",
		MerchantId: uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, vld, service)

	router := chi.NewRouter()
	router.Get("/insights", handler.GetPayoutInsights)

	testCases := []struct {
		name             string
		queryParams      string
		userClaim        *userModel.UserTokenClaims
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR: User not in context",
			queryParams:      "",
			userClaim:        nil,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On("GetPayoutInsights", mock.Anything, mock.Anything).
					Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid startDate format",
			queryParams:      "?startDate=invalid-date",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"parsing time \"invalid-date\" as \"2006-01-02T15:04:05Z\": cannot parse \"invalid-date\" as \"2006\"","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid endDate format",
			queryParams:      "?endDate=invalid-date",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"parsing time \"invalid-date\" as \"2006-01-02T15:04:05Z\": cannot parse \"invalid-date\" as \"2006\"","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: No waiting payouts",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On("GetPayoutInsights", mock.Anything, mock.Anything).
					Once().Return(&model.GetPayoutInsightsResponse{
					TotalTransaction: 0,
					TotalAmount: commonModel.Amount{
						Currency: "IDR",
						Value:    "0.00",
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"totalAmount":{"currency":"IDR","value":"0.00"},"totalTransaction":0}}`,
		},
		{
			name:        "SUCCESS: With waiting payouts",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On("GetPayoutInsights", mock.Anything, mock.Anything).
					Once().Return(&model.GetPayoutInsightsResponse{
					TotalTransaction: 3,
					TotalAmount: commonModel.Amount{
						Currency: "IDR",
						Value:    "5000000.00",
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"totalAmount":{"currency":"IDR","value":"5000000.00"},"totalTransaction":3}}`,
		},
		{
			name:        "SUCCESS: With date filter",
			queryParams: "?startDate=2026-03-01T00:00:00Z&endDate=2026-03-31T23:59:59Z",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On("GetPayoutInsights", mock.Anything, mock.Anything).
					Once().Return(&model.GetPayoutInsightsResponse{
					TotalTransaction: 2,
					TotalAmount: commonModel.Amount{
						Currency: "IDR",
						Value:    "3000000.00",
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"totalAmount":{"currency":"IDR","value":"3000000.00"},"totalTransaction":2}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/insights"+tt.queryParams, nil)

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
