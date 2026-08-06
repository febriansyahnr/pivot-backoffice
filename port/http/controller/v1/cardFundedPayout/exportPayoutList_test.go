package cardFundedPayoutController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportPayoutList(t *testing.T) {
	vld := validatorExt.New()

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe",
		MerchantId: uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}

	testCases := []struct {
		name             string
		requestBody      any
		userClaim        *userModel.UserTokenClaims
		timezoneHeader   string
		setupMock        func(*serviceMocks.ICardFundedPayoutService)
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:           "SUCCESS",
			requestBody:    model.FilterGetPayoutList{},
			userClaim:      validUserClaims,
			timezoneHeader: "Asia/Jakarta",
			setupMock: func(service *serviceMocks.ICardFundedPayoutService) {
				service.On(
					"ExportPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&model.ExportPayoutListResponse{
					Url: "https://storage.example.com/exports/payout-list.xlsx",
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"url":"https://storage.example.com/exports/payout-list.xlsx"}}`,
		},
		{
			name:             "ERROR: User not in Context",
			requestBody:      model.FilterGetPayoutList{},
			userClaim:        nil,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid JSON body",
			requestBody:      "invalid json",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid character 'i' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Service error",
			requestBody:    model.FilterGetPayoutList{},
			userClaim:      validUserClaims,
			timezoneHeader: "Asia/Jakarta",
			setupMock: func(service *serviceMocks.ICardFundedPayoutService) {
				service.On("ExportPayoutList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Without timezone header",
			requestBody: model.FilterGetPayoutList{},
			userClaim:   validUserClaims,
			setupMock: func(service *serviceMocks.ICardFundedPayoutService) {
				service.On(
					"ExportPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&model.ExportPayoutListResponse{
					Url: "https://storage.example.com/exports/payout-no-tz.xlsx",
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"url":"https://storage.example.com/exports/payout-no-tz.xlsx"}}`,
		},
		{
			name:           "SUCCESS: With UTC timezone",
			requestBody:    model.FilterGetPayoutList{},
			userClaim:      validUserClaims,
			timezoneHeader: "UTC",
			setupMock: func(service *serviceMocks.ICardFundedPayoutService) {
				service.On(
					"ExportPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&model.ExportPayoutListResponse{
					Url: "https://storage.example.com/exports/payout-utc.xlsx",
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"url":"https://storage.example.com/exports/payout-utc.xlsx"}}`,
		},
		{
			name:           "SUCCESS: Empty URL in response",
			requestBody:    model.FilterGetPayoutList{},
			userClaim:      validUserClaims,
			timezoneHeader: "Asia/Jakarta",
			setupMock: func(service *serviceMocks.ICardFundedPayoutService) {
				service.On(
					"ExportPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&model.ExportPayoutListResponse{
					Url: "",
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"url":""}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			service := serviceMocks.NewICardFundedPayoutService(t)
			handler := New(cfg, vld, service)

			router := chi.NewRouter()
			router.Post("/", handler.ExportPayoutList)

			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			var body bytes.Buffer
			if tt.requestBody != nil {
				switch v := tt.requestBody.(type) {
				case string:
					body.WriteString(v)
				default:
					_ = json.NewEncoder(&body).Encode(v)
				}
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", &body)

			if tt.timezoneHeader != "" {
				req.Header.Set(constant.HeaderTimeZoneKey, tt.timezoneHeader)
			}

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
