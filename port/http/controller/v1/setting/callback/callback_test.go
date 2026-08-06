package callback_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/callback"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPerformCallbackSettingReqFunc(t *testing.T) {
	validator := validatorExt.New()
	service := serviceMocks.NewICallbackService(t)

	callbackPath := "/settings/callbacks"
	callbackApiKeyPath := "/settings/callbacks/api-key"

	handler := New(validator, nil, service)

	router := chi.NewRouter()
	router.Get(callbackPath, handler.Get)
	router.Get(callbackApiKeyPath, handler.GetApiKey)

	defPath := callbackPath
	userClaims := &user.UserTokenClaims{
		UUID:       "46d63425-a899-4f0e-a7f4-3bec11dc4420",
		MerchantId: "b73a8ba5-6d80-4da0-b1a0-e694ef3b75b7",
	}
	callbackSetReqMockType := mock.AnythingOfType("*callback_model.CallbackURLSettingReq")

	tests := []struct {
		name           string
		path           string
		userClaims     *user.UserTokenClaims
		setupReq       func(r *http.Request)
		setupMocks     func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Unauthorized user",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:           "ERROR:Invalid request body",
			userClaims:     &user.UserTokenClaims{UUID: userClaims.UUID},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[{"field":"MerchantID","message":"Key: 'CallbackURLSettingReq.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}],"traceId":"","type":"API_ERROR"}, "message":"invalid validation"}`,
		},
		{
			name:       "ERROR:View callback api key",
			path:       callbackApiKeyPath,
			userClaims: userClaims,
			setupMocks: func() {
				service.On(
					"GetCallbackAPIKeyByMerchantId", constant.ValueCtxMockType(), callbackSetReqMockType,
				).Once().Return(nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPIN))
			},
			setupReq: func(r *http.Request) {
				r.Header.Set(constant.HeaderXRequestPIN, "not-encoded-base64")
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid pin"}`,
		},
		{
			name:       "SUCCESS:Callback dashboard",
			path:       callbackPath,
			userClaims: userClaims,
			setupMocks: func() {
				service.On(
					"GetCallbackURLByMerchantId", constant.ValueCtxMockType(), callbackSetReqMockType,
				).Once().Return([]callbackModel.CallbackURLSettingResp{
					{
						MasterID:           "1",
						MasterName:         "2",
						CallbackID:         constant.NullString{NullString: sql.NullString{String: "3", Valid: true}},
						CallbackURL:        constant.NullString{NullString: sql.NullString{String: "4", Valid: true}},
						CallbackBaseURL:    constant.NullString{NullString: sql.NullString{String: "5", Valid: true}},
						CallbackLastUpdate: constant.NullTime{NullTime: sql.NullTime{Time: time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC), Valid: true}},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":[{"masterId":"1","masterName":"2","callbackId":"3","callbackUrl":"4","callbackBaseURL":"5","callbackLastUpdate":"2024-06-13T00:00:00Z","callbackTemplate":{"event":"","data":null}}]}`,
		},
		{
			name:       "SUCCESS:View callback api key",
			path:       callbackApiKeyPath,
			userClaims: userClaims,
			setupMocks: func() {
				service.On(
					"GetCallbackAPIKeyByMerchantId", constant.ValueCtxMockType(), callbackSetReqMockType,
				).Once().Return(callbackModel.CallbackAPIKeyResp{APIKey: "unique-api-key"}, nil)
			},
			setupReq: func(r *http.Request) {
				r.Header.Set(constant.HeaderXRequestPIN, "MTIzNDU2")
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"apiKey":"unique-api-key"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.path == "" {
				test.path = defPath
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, test.path, nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}
			if test.setupReq != nil {
				test.setupReq(req)
			}
			if test.setupMocks != nil {
				test.setupMocks()
			}
			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
