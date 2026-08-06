package xbPayoutController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentDashboardInsights(t *testing.T) {
	xbPayoutSvc := mockService.NewIXbPayoutService(t)

	handler := New(nil, WithXbPayoutService(xbPayoutSvc))

	router := chi.NewRouter()
	router.Get("/xb-payout-insights", handler.GetXbPayoutDashboardInsights)

	now := time.Now().UTC()
	userClaims := &user.UserTokenClaims{}
	queryParams := fmt.Sprintf("insightStartDate=%s&insightEndDate=%s", now.AddDate(0, 0, -1).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))

	tests := []struct {
		name             string
		claims           *user.UserTokenClaims
		params           string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:User not found", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:             "ERROR:Invalid date range format", // NOSONAR
			claims:           userClaims,
			params:           "insightStartDate=ABC&insightEndDate=DEF", // NOSONAR
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "Key: insightStartDate Value: ABC Error: Value format must be yyyy-mm-ddThh:nn:ssZ"),
		},
		{
			name:             "ERROR:Empty date range", // NOSONAR
			claims:           userClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "start date or end date cannot be empty"),
		},
		{
			name:   "ERROR:Some error", // NOSONAR
			claims: userClaims,
			params: queryParams,
			setupMock: func() {
				xbPayoutSvc.On("GetXbPayoutDashboardInsights", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: constant.WrapErrApiRespForTest(99, response.ErrTypeAPI, "an error occurred on the server. please try again later"),
		},
		{
			name:   "SUCCESS", // NOSONAR
			claims: userClaims,
			params: queryParams,
			setupMock: func() {
				xbPayoutSvc.On("GetXbPayoutDashboardInsights", mock.Anything, mock.Anything).Once().Return(&disbursementModel.XbPayoutDashboardInsights{
					WaitingForConfirmCount:    1, // NOSONAR
					InformationRequestedCount: 2, // NOSONAR
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"waitingForConfirmCount":1,"informationRequestedCount":2,"pendingCount":0,"successCount":0,"successTotal":0.00,"topCountriesByVolume":null}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/xb-payout-insights?"+test.params, nil)

			if test.claims != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), constant.CtxUserInfoKey, test.claims),
				)
			}
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Result:", rec.Body.String())
			}
		})
	}
}
