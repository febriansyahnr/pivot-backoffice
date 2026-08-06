package ipWhitelistController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {

	testCases := []struct {
		name         string
		expectedCode int
		expectedBody string
		setup        func(svc *mockSvc.IIPWhitelistService)
		setupClaims  bool
		setupId      bool
	}{
		{
			name: "SUCCESS: Create IP Configuration",
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"Delete",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			setupClaims:  true,
			setupId:      true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"OK","data":null}`,
		},
		{
			name: "ERROR: Unable get claims",
			setup: func(svc *mockSvc.IIPWhitelistService) {

			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			setupClaims:  false,
			setupId:      true,
		},
		{
			name: "ERROR: No ID",
			setup: func(svc *mockSvc.IIPWhitelistService) {

			},
			setupClaims:  true,
			setupId:      false,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid id","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Delete IP Configuration",
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"Delete",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("errors"))
			},
			setupId:      true,
			setupClaims:  true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIIPWhitelistService(t)
			mockValidator := validator.New()
			tc.setup(svc)

			ctrl := New(svc, mockValidator)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			if tc.setupClaims {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: uuid.NewString(),
				})
				req = req.WithContext(ctx)
			}

			if tc.setupId {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("id", uuid.NewString())
				req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, routeCtx))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.Delete)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())

		})
	}
}
