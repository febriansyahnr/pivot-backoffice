package vendor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	testCases := []struct {
		name         string
		paramID      string
		setup        func(svc *mockSvc.IVendorService)
		expectedCode int
		expectedBody string
	}{
		{
			name:    "SUCCESS: Delete Vendor",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Delete",
					mock.Anything,
					"123e4567-e89b-12d3-a456-426614174000",
				).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":null}`,
		},
		{
			name:         "ERROR: Invalid UUID",
			paramID:      "not-a-valid-uuid",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid id",
				"error": {
					"type": "API_ERROR",
					"message": "invalid id",
					"recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:    "ERROR: Service Error",
			paramID: "123e4567-e89b-12d3-a456-426614174999",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On(
					"Delete",
					mock.Anything,
					"123e4567-e89b-12d3-a456-426614174999",
				).Return(errors.New("not found"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "not found",
				"error": {
					"type": "UNKNOWN",
					"message": "not found",
					"recommendation": ""
				},
				"data": null
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIVendorService(t)

			tc.setup(svc)

			ctrl := New(svc, nil)

			req := httptest.NewRequest(http.MethodDelete, "/card-funded-payout/vendors/"+tc.paramID, nil)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Delete)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
