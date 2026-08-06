package withdrawalController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWithdrawalInsight(t *testing.T) {
	var (
		validUserClaims = &user.UserTokenClaims{
			UUID:       uuid.NewString(),
			MerchantId: uuid.NewString(),
		}

		mockWithdrawalService mockService.IWithdrawalService
		insightController     = New(nil, nil, &mockWithdrawalService, nil)
		defaultInsight        = &withdrawal.WithdrawalInsightItem{
			Total: 0,
			TotalAmount: commonModel.Amount{
				Currency: "IDR",
				Value:    strconv.FormatFloat(0, 'f', 2, 64),
			},
		}
	)

	testCases := []struct {
		name           string
		callMock       func()
		expectedStatus int
		userClaim      *user.UserTokenClaims
		response       string
	}{
		{
			name:      "when everything is ok, should return 200",
			userClaim: validUserClaims,
			callMock: func() {
				mockWithdrawalService.On("GetTodayWithdrawalInsight", mock.Anything, withdrawal.WithdrawalInsightRequest{
					MerchantID: validUserClaims.MerchantId,
				}).
					Return(&withdrawal.WithdrawalInsightResponse{
						TodayTotalSuccess: defaultInsight,
						TodayTotalPending: defaultInsight,
						TodayTotalFailed:  defaultInsight,
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			response:       `{"code":"00","message":"OK","data":{"todayTotalSuccess":{"total":0,"totalAmount":{"currency":"IDR","value":"0.00"}},"todayTotalPending":{"total":0,"totalAmount":{"currency":"IDR","value":"0.00"}},"todayTotalFailed":{"total":0,"totalAmount":{"currency":"IDR","value":"0.00"}}}}`,
		},
		{
			name:      "when error occurred on get insight, then should return 500",
			userClaim: validUserClaims,
			callMock: func() {
				mockWithdrawalService.On("GetTodayWithdrawalInsight", mock.Anything, withdrawal.WithdrawalInsightRequest{
					MerchantID: validUserClaims.MerchantId,
				}).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("database error"))).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			response:       constant.WrapErrApiRespForTest(99, response.ErrTypeAPI, "database error"),
		},
		{
			name:           "when request does't have auth header, then should return 401",
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
			response:       constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/api/v1/withdrawals/insights", nil)
			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(insightController.GetInsights)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tc.response, rr.Body.String()) {
				t.Log("Result:", rr.Body.String())
			}
		})
	}
}
