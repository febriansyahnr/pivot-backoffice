package depositSettingController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/deposit"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {

	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(nil, nil, merchantSvc)

	router := chi.NewRouter()
	router.Get("/settings/deposit", handler.Get)

	userClaims := &user.UserTokenClaims{
		MerchantId: "0192d7c3-d4cb-74b6-a1ec-5c08566124c6",
	}

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:       "ERROR:Some error", // NOSONAR
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"GetDepositSetting", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:       "SUCCESS", // NOSONAR
			userClaims: userClaims,
			setupMock: func() {
				merchantSvc.On(
					"GetDepositSetting", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.DepositSettingResponse{
					MerchantName:   "Test",
					AutoWithdrawal: "ON",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"merchantName":"Test", "autoWithdrawal":"ON"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/settings/deposit", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}
			req.Header.Set(c.HeaderXSubMerchantID, "1fd6cdc2-563b-4583-832b-987e95aff71b")

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
