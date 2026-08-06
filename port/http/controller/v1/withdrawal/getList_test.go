package withdrawalController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), logger, service, nil)

	router := chi.NewRouter()
	router.Get("/withdrawals/{account}", handler.GetList)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	loc, _ := time.LoadLocation(c.TimeLoc)

	now := time.Now().In(loc)
	endDateStr := now.Format(time.DateOnly)
	startDateStr := now.AddDate(0, 0, -14).Format(time.DateOnly)

	tests := []struct {
		name            string
		account         string // default if empty is payments
		userTokenClaims *user.UserTokenClaims
		queries         string
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
			name:            "ERROR:Invalid start date format",
			userTokenClaims: userTokenClaims,
			queries:         "startDate=ABCV",
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid start date format"),
		},
		{
			name:            "ERROR:Invalid end date format",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=ABC", startDateStr),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid end date format"),
		},
		{
			name:            "ERROR:Invalid filter date input",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s", startDateStr, now.AddDate(0, -2, 0).Format(time.DateOnly)),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "startDate must not be greater than endDate"),
		},
		{
			name:            "ERROR:Time span exceeding 31 days",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s", startDateStr, now.AddDate(0, 0, 30).Format(time.DateOnly)),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "The date range exceeds the allowed limit. Maximum permitted is 31 days."),
		},
		{
			name:            "ERROR:Invalid page number format",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s&page=`1`", startDateStr, endDateStr),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid page format. Use number format instead"),
		},
		{
			name:            "ERROR:Invalid per page format",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s&page=1&perPage=`10`", startDateStr, endDateStr),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid perPage format. Use number format instead"),
		},
		{
			name:            "ERROR:Invalid transaction status",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s&page=1&perPage=10&status=INIT", startDateStr, endDateStr),
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Status","message":"Key: 'WithdrawalHistoryRequest.WithdrawalListRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:            "ERROR:Some error", // NOSONAR
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s", startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"GetList", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalHistoryRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS:Default filter",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s", startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"GetList", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalHistoryRequest"),
				).Return(&commonModel.PaginationResponse{
					Data: []withdrawal.WithdrawalHistoryResponse{{
						Id:                     "1",
						Type:                   "6",
						Amount:                 2,
						BeneficiaryBankName:    "3",
						BeneficiaryAccountName: "4",
						Status:                 "5",
					}},
					Meta: commonModel.Meta{
						Page: 6, PerPage: 7, TotalItems: 8, TotalPages: 9,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"id":"1","date":"0001-01-01T00:00:00Z","type":"6","amount":2,"beneficiaryBankName":"3","beneficiaryAccountName":"4","status":"5","createdBy":"","balanceType":""}],"pagination":{"page":6,"perPage":7,"totalItems":8,"totalPages":9}}`,
		},
		{
			name:            "SUCCESS:Custom filter",
			userTokenClaims: userTokenClaims,
			queries:         fmt.Sprintf("startDate=%s&endDate=%s&id=3cd7c41a-1871-4edb-b652-2eb67d74d7b1&perPage=1000", startDateStr, endDateStr),
			wantStatusCode:  http.StatusOK,
			wantRespBody:    `{"code":"00","message":"OK","data":[{"id":"1","date":"0001-01-01T00:00:00Z","type":"6","amount":2,"beneficiaryBankName":"3","beneficiaryAccountName":"4","status":"5","createdBy":"","balanceType":""}],"pagination":{"page":6,"perPage":7,"totalItems":8,"totalPages":9}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.account == "" {
				test.account = "payments"
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/withdrawals/"+test.account, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}
			req.URL.RawQuery = test.queries

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
