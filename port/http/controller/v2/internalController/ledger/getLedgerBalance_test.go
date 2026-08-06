package ledgerController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLedgerBalance(t *testing.T) {
	testCases := []struct {
		name             string
		expectedCode     int
		expectedResponse string
		setup            func(mockService *mockSvc.ILedgerService)
		setupParam       func(*chi.Context)
	}{
		{
			name:             "SUCCESS",
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"Balance":992,"Currency":"IDR"}}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("GetLedgerBalance", mock.Anything, mock.Anything).Return(&ledger_model.LedgerBalance{Balance: 992, Currency: constant.CurrencyIDR}, nil)
			},
			setupParam: func(param *chi.Context) {
				param.URLParams.Add("accountId", uuid.New().String())
			},
		},
		{
			name:             "ERROR: No Account ID",
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"invalid UUID length: 0","error":{"type":"API_ERROR","message":"invalid UUID length: 0","recommendation":""},"data":null}`,
			setup: func(mockService *mockSvc.ILedgerService) {
			},
			setupParam: func(param *chi.Context) {
			},
		},
		{
			name:             "ERROR: Get balance",
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("GetLedgerBalance", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			setupParam: func(param *chi.Context) {
				param.URLParams.Add("accountId", uuid.New().String())
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockSvc.ILedgerService)
			tc.setup(mockService)
			ctrl := New(mockService)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/url", nil)
			chiRouterCtx := chi.NewRouteContext()
			tc.setupParam(chiRouterCtx)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx)
			req = req.WithContext(ctx)

			ctrl.GetLedgerBalance(w, req)
			assert.Equal(t, tc.expectedCode, w.Code)
			assert.JSONEqf(t, tc.expectedResponse, w.Body.String(), "want: %v, got: %v", tc.expectedResponse, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
