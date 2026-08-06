package orchestratorController_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/orchestrator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDetailById(t *testing.T) {
	transactionId := uuid.NewString()
	claims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	handler := New(&config.Config{}, orchestratorSvc, nil, nil, nil)

	router := chi.NewRouter()
	router.Get("/details/{transaction_id}", handler.GetDetailById)

	wrapErrResp := func(code int, errType, msg string) string {
		return fmt.Sprintf(`{"code":"%d","data":null,"error":{"details":[],"traceId":"","type":"%s"},"message":"%s"}`, code, errType, msg)
	}
	response := &orchestratorModel.TransactionHistoryDetailResp{
		Id:        transactionId,
		CreatedAt: time.Now().UTC(),
		Type:      constant.TypeDisbursement,
		Remarks:   "Test",
		Amount:    10_000,
		Status:    constant.StatusSuccess,
	}
	rawResponse, err := json.Marshal(response)

	require.NoError(t, err)

	tests := []struct {
		name           string
		transactionId  string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid format transaction id",
			transactionId:  "abc",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrResp(40, "API_ERROR", "invalid transaction id format"),
		},
		{
			name:           "ERROR:User not found",
			transactionId:  transactionId,
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrResp(41, "API_ERROR", "user not found"),
		},
		{
			name:          "ERROR:Get detail transaction",
			transactionId: transactionId,
			userClaims:    claims,
			setupMock: func() {
				orchestratorSvc.On(
					"GetDetailById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrResp(99, "UNKNOWN", "some error"),
		},
		{
			name:          "SUCCESS",
			transactionId: transactionId,
			userClaims:    claims,
			setupMock: func() {
				orchestratorSvc.On(
					"GetDetailById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(response, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":%s}`, string(rawResponse)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/details/"+test.transactionId, nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}
			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
