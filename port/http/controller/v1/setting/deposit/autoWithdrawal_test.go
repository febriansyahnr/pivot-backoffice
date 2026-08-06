package depositSettingController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/deposit"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestSetAutoWithdrawal(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(validatorExt.New(), nil, merchantSvc)

	router := chi.NewRouter()
	router.Patch("/settings/deposit/auto-withdrawal", handler.SetAutoWithdrawal)

	userClaims := &user.UserTokenClaims{
		UUID:       "0192d7c3-d4cb-74b6-a1ec-5c08566124c7",
		MerchantId: "0192d7c3-d4cb-74b6-a1ec-5c08566124c6",
	}
	request := &merchant.AutoWithdrawalSettingRequest{
		UserId:     userClaims.UUID,
		MerchantId: "1fd6cdc2-563b-4583-832b-987e95aff71b",
		Status:     "OFF",
	}
	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		requestBody    string
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
			name:           "ERROR:Invalid request body",
			userClaims:     userClaims,
			requestBody:    `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR:Empty status",
			userClaims:     userClaims,
			requestBody:    `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Status","message":"Key: 'AutoWithdrawalSettingRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			userClaims:  userClaims,
			requestBody: `{"status":"OFF","MerchantId":"ABC","merchantId":"CBA"}`,
			setupMock: func() {
				merchantSvc.On(
					"SetAutoWithdrawal", c.ValueCtxMockType(), request,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:        "SUCCESS", // NOSONAR
			userClaims:  userClaims,
			requestBody: `{"status":"OFF"}`,
			setupMock: func() {
				merchantSvc.On("SetAutoWithdrawal", c.ValueCtxMockType(), request).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"status":"OFF"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/settings/deposit/auto-withdrawal", strings.NewReader(test.requestBody))

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
