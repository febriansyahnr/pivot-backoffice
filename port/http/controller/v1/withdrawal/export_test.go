package withdrawalController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExport(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), nil, service, nil)

	router := chi.NewRouter()
	router.Post("/withdrawals/{account}/export", handler.Export)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "000c4096-e92e-4f59-a0fe-bab1fd53b1c9",
		MerchantId: "23bd62a9-239f-4d90-a6b0-6bf22fdec793",
	}

	loc, _ := time.LoadLocation(c.TimeLoc)

	now := time.Now().In(loc)
	endDateStr := now.Format(time.DateOnly)
	startDateStr := now.AddDate(0, 0, -14).Format(time.DateOnly)

	tests := []struct {
		name            string
		account         string // default if empty is payments
		userTokenClaims *user.UserTokenClaims
		requestBody     string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR:Invalid path URL",
			account:        "invalid",
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   c.WrapErrApiRespForTest(44, s.ErrTypeAPI, "invalid path URL"),
		},
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR:Invalid request body",
			userTokenClaims: userTokenClaims,
			requestBody:     `A`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:            "ERROR:Invalid start date format",
			userTokenClaims: userTokenClaims,
			requestBody:     `{"startDate": "ABCV"}`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid start date format"),
		},
		{
			name:            "ERROR:Invalid end date format",
			userTokenClaims: userTokenClaims,
			requestBody:     `{"startDate": "2024-09-01", "endDate": "ABCV"}`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid end date format"),
		},
		{
			name:            "ERROR:Invalid filter date input",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, now.AddDate(0, 0, -20).Format(time.DateOnly)),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "startDate must not be greater than endDate"),
		},
		{
			name:            "ERROR:Time span exceeding 31 days",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, now.AddDate(0, 0, 30).Format(time.DateOnly)),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "The date range exceeds the allowed limit. Maximum permitted is 31 days."),
		},
		{
			name:            "ERROR:Invalid transaction status",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s", "status": "INIT"}`, startDateStr, endDateStr),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Status","message":"Key: 'WithdrawalListRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:            "ERROR:Some error", // NOSONAR
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"Export", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"Export", c.ValueCtxMockType(), mock.Anything,
				).Return(&withdrawal.WithdrawalDownloadResponse{URL: "https://"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"url":"https://"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.account == "" {
				test.account = "payments"
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/withdrawals/"+test.account+"/export", strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
