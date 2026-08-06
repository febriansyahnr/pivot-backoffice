package orchestratorController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/orchestrator"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportToExcelTransactionHistory(pt *testing.T) {
	service := mockServices.NewIOrchestratorService(pt)

	router := chi.NewRouter()
	router.Get(
		"/exports", New(&config.Config{}, service, nil, validator.New(), nil).ExportToExcelTransactionHistory,
	)

	now := time.Now().UTC()
	endDateStr := now.Format(time.DateOnly)
	startDateStr := now.AddDate(0, 0, -10).Format(time.DateOnly)
	queryParams := fmt.Sprintf("startSettlementDate=%s&endSettlementDate=%s", startDateStr, endDateStr)

	tests := []struct {
		name            string
		queries         string
		userClaims      *user.UserTokenClaims
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
		wantRespHeaders func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "ERROR:Unauthorized",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, "API_ERROR", "user not found"),
		},
		{
			name:           "ERROR:Empty filter date",
			userClaims:     &user.UserTokenClaims{},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, "API_ERROR", "start settlement date and end settlement date must be filled"),
		},
		{
			name:           "ERROR:Max date range",
			queries:        "startSettlementDate=2025-06-30&endSettlementDate=2025-07-31",
			userClaims:     &user.UserTokenClaims{},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, "API_ERROR", "maximum date range 31 days"),
		},
		{
			name:       "ERROR:Some error",
			queries:    queryParams,
			userClaims: &user.UserTokenClaims{},
			setupMock: func() {
				service.On(
					"GenExcelForTransactionHistories", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, "UNKNOWN", "some error"),
		},
		{
			name:       "SUCCESS",
			queries:    queryParams,
			userClaims: &user.UserTokenClaims{},
			setupMock: func() {
				service.On(
					"GenExcelForTransactionHistories", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					_, _ = args.Get(1).(*model.FileWriter).Write([]byte(`{"message":"OK"}`))
				}).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`, // NOSONAR
			wantRespHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, rec.Result().Header.Get(c.HeaderDataOrigin), c.DataOriginRaw)
			},
		},
		{
			name:    "SUCCESS:Data reporting",
			queries: queryParams,
			userClaims: &user.UserTokenClaims{
				MerchantId: "79f52ad0-4820-46fd-8d38-10e5a0a8514c", // NOSONAR
			},
			setupMock: func() {
				service.On(
					"GenExcelForTransactionHistories", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					_, _ = args.Get(1).(*model.FileWriter).Write([]byte(`{"message":"OK"}`))
				}).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`, // NOSONAR
			wantRespHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, rec.Result().Header.Get(c.HeaderDataOrigin), c.DataOriginReporting)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/exports", nil)
			req.URL.RawQuery = test.queries

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}
			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			if test.wantRespHeaders != nil {
				test.wantRespHeaders(t, rec)
			}

			service.AssertExpectations(t)
		})
	}
}
