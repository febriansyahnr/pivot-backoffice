package accountController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetByUUID(pt *testing.T) {
	const requestID = "e39d1d21-0a7b-49e5-8cd9-404ac75d54be"
	tests := []struct {
		name           string
		requestID      string
		mockSetup      func(b *mockServices.IAccountService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:      "ERROR:Invalid request ID format",
			requestID: "invalid",
			mockSetup: func(b *mockServices.IAccountService) {
				// Empty mock setup
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"invalid UUID length: 7"}`,
		},
		{
			name:      "ERROR:Invalid db session",
			requestID: requestID,
			mockSetup: func(b *mockServices.IAccountService) {
				b.On(
					"GetAccount", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(),
				).Once().Return(nil, errors.New("invalid db session"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"invalid db session"}`,
		},
		{
			name:      "ERROR:Balance not found",
			requestID: requestID,
			mockSetup: func(b *mockServices.IAccountService) {
				b.On(
					"GetAccount", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(),
				).Once().Return(nil, nil)
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   `{"code":"44","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"data not found"}`,
		},
		{
			name:      "SUCCESS",
			requestID: requestID,
			mockSetup: func(b *mockServices.IAccountService) {
				b.On(
					"GetAccount", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(),
				).Return(func() (*accountModel.Account, error) {
					r := &accountModel.Account{}
					r.UUID, _ = uuid.Parse(requestID)
					r.ReferenceID = r.UUID
					r.Name = "John Wick"
					return r, nil
				}())
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"e39d1d21-0a7b-49e5-8cd9-404ac75d54be","merchantId":"e39d1d21-0a7b-49e5-8cd9-404ac75d54be","name":"John Wick","type":"","eodBalance":0,"currency":"","lastUpdateBalanceAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","current_balance_check_time":"0001-01-01T00:00:00Z","userType":""}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			mockAccount := mockServices.NewIAccountService(pt)
			test.mockSetup(mockAccount)

			mc := New(nil, mockAccount, nil)
			req := httptest.NewRequest(http.MethodGet, "/accounts/"+test.requestID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.requestID)

			rec := httptest.NewRecorder()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler := http.HandlerFunc(mc.GetByUUID)
			handler.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
