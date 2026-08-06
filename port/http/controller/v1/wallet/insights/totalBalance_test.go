package walletInsightsController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/walletInsights"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTotalBalance(t *testing.T) {

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: "3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
	}

	tests := []struct {
		name           string
		userClaim      *user.UserTokenClaims
		setupMock      func(*mocks.IWalletInsightService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:      "SUCCESS",
			userClaim: userTokenClaims,
			setupMock: func(svc *mocks.IWalletInsightService) {

				svc.On("TotalBalance", mock.Anything, mock.Anything, mock.Anything).Return(&walletInsights.MerchantTotalBalance{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"totalBalance":0,"lastUpdatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:      "ERROR: Invalid Claims",
			userClaim: nil,
			setupMock: func(svc *mocks.IWalletInsightService) {
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR",
			userClaim: userTokenClaims,
			setupMock: func(svc *mocks.IWalletInsightService) {

				svc.On("TotalBalance", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/total-balance", nil)

			svc := mocks.NewIWalletInsightService(t)

			handler := New(svc)

			router := chi.NewRouter()
			router.Post("/total-balance", handler.TotalBalance)

			if test.setupMock != nil {
				test.setupMock(svc)
			}
			if test.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaim))
			}
			q := req.URL.Query()
			q.Add("refresh", "true")
			req.URL.RawQuery = q.Encode()

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
