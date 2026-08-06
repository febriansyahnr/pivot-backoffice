package accountController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBalance(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleApprover,
	}
	errResponseBody := `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`

	testCases := []struct {
		name           string
		userClaim      *user.UserTokenClaims
		query          string
		setupMock      func(*serviceMocks.IOrchestratorService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Unauthorized - user not found in context",
			userClaim:      nil,
			setupMock:      func(*serviceMocks.IOrchestratorService) { /* Empty Function */ },
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   errResponseBody,
		},
		{
			name:           "ERROR: Invalid currency",
			userClaim:      validUserClaims,
			query:          "usecase=DISBURSEMENT&currency=IDK",
			setupMock:      func(*serviceMocks.IOrchestratorService) { /* Empty Function */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   errResponseBody,
		},
		{
			name:      "ERROR: GetMerchantBalance service returns error",
			userClaim: validUserClaims,
			query:     "usecase=DISBURSEMENT",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("GetMerchantBalance",
					mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   errResponseBody,
		},
		{
			name:      "SUCCESS: default usecase (disbursement)",
			userClaim: validUserClaims,
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("GetMerchantBalance",
					mock.Anything, mock.MatchedBy(func(req orchestratorModel.GetMerchantBalanceRequest) bool {
						return req.BalanceName == constant.TypeDisbursement && req.MerchantID == validUserClaims.MerchantId
					}),
				).Once().Return(&orchestratorModel.GetMerchantBalanceResponse{
					AvailableBalance: commonModel.Amount{Currency: "IDR", Value: "10000.00"},
					PendingBalance:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"availableBalance":{"currency":"IDR","value":"10000.00"},"pendingBalance":{"currency":"IDR","value":"0.00"}},"message":"Success"}`,
		},
		{
			name:      "SUCCESS: payment usecase",
			userClaim: validUserClaims,
			query:     "usecase=PAYMENT",
			setupMock: func(orchestratorSvc *serviceMocks.IOrchestratorService) {
				orchestratorSvc.On("GetMerchantBalance",
					mock.Anything, mock.MatchedBy(func(req orchestratorModel.GetMerchantBalanceRequest) bool {
						return req.BalanceName == constant.TypePayment
					}),
				).Once().Return(&orchestratorModel.GetMerchantBalanceResponse{
					AvailableBalance: commonModel.Amount{Currency: "IDR", Value: "50000.00"},
					PendingBalance:   commonModel.Amount{Currency: "IDR", Value: "25000.00"},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"availableBalance":{"currency":"IDR","value":"50000.00"},"pendingBalance":{"currency":"IDR","value":"25000.00"}},"message":"Success"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			handler := New(nil, nil, orchestratorSvc)

			rec := httptest.NewRecorder()

			url := "/balances"
			if tc.query != "" {
				url += "?" + tc.query
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, uuid.NewString())
			if tc.userClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, tc.userClaim)
			}
			req = req.WithContext(ctx)

			router := chi.NewRouter()
			router.Get("/balances", handler.GetBalance)
			tc.setupMock(orchestratorSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})
	}
}
