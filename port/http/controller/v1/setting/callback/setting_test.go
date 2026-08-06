package callback_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	callbackPartnerModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/callback"

	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestAndSaveCallbackURL(t *testing.T) {
	validator := validatorExt.New()
	service := serviceMocks.NewICallbackService(t)

	handler := New(validator, nil, service)

	router := chi.NewRouter()
	router.Post("/callbacks/urls/{master_id}", handler.TestAndSaveCallbackURL)

	masterID := uuid.NewString()
	userClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}
	payload := &callbackModel.TestAndSaveCallbackURLReq{
		Name: "PAYMENT",
		URL:  "https://example.id/payment/notification",
		Payload: callbackPartnerModel.CallbackPayloadRequest{
			Event: "PAYOUT.TEST",
			Data:  map[string]string{"test": "OK"},
		},
	}
	response := &callbackModel.TestAndSaveCallbackURLResp{
		Status: true,
		Information: callbackModel.CallbackURLInfoResp{
			Product: "PAYMENT",
			Event:   "PAYOUT.TEST",
		},
		Body:      map[string]string{"test": "OK"},
		Duration:  "2s",
		RequestID: "unique-request-id",
	}
	responseBody, _ := json.Marshal(response)

	requestMockType := mock.AnythingOfType("*callback_model.TestAndSaveCallbackURLReq")

	tests := []struct {
		name           string
		masterId       string
		request        string
		userClaims     *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			masterId:       "00000",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:           "ERROR:Invalid body format",
			userClaims:     userClaims,
			request:        `{\}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid character '\\\\' looking for beginning of object key string"}`,
		},
		{
			name:           "ERROR:Invalid request data",
			masterId:       "0000",
			userClaims:     userClaims,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","data":null,"error":{"details":[{"field":"CallbackMasterID","message":"Key: 'TestAndSaveCallbackURLReq.CallbackMasterID' Error:Field validation for 'CallbackMasterID' failed on the 'uuid' tag"}],"traceId":"","type":"API_ERROR"},"message":"invalid validation"}`,
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"TestAndSaveCallbackURL", c.ValueCtxMockType(), requestMockType,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name:       "SUCCESS",
			userClaims: userClaims,
			setupMock: func() {
				service.On(
					"TestAndSaveCallbackURL", c.ValueCtxMockType(), requestMockType,
				).Return(response, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":%s}`, string(responseBody)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.setupMock != nil {
				test.setupMock()
			}

			if test.masterId == "" {
				test.masterId = masterID
			}
			if test.request == "" {
				buf, _ := json.Marshal(payload)
				test.request = string(buf)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/callbacks/urls/"+test.masterId, strings.NewReader(test.request))

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
