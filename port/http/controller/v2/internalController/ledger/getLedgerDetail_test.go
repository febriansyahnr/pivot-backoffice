package ledgerController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLedgerTeDetail(t *testing.T) {
	testCases := []struct {
		name             string
		expectedCode     int
		expectedResponse string
		setup            func(mockService *mockSvc.ILedgerService)
	}{
		{
			name:             "success",
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":[]}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("GetLedgerDetail", mock.Anything, mock.Anything).Return([]orchestrator_model.AccountTransaction{}, nil)
			},
		},
		{
			name:             "error",
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("GetLedgerDetail", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
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
			chiRouterCtx.URLParams.Add("referenceId", uuid.New().String())
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx)
			req = req.WithContext(ctx)

			ctrl.GetLedgerDetail(w, req)
			assert.Equal(t, tc.expectedCode, w.Code)
			assert.JSONEqf(t, tc.expectedResponse, w.Body.String(), "want: %v, got: %v", tc.expectedResponse, w.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
