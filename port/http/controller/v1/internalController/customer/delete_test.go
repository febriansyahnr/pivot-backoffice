package customerController

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	uuid := uuid.New().String()
	testCases := []struct {
		name               string
		setup              func(customerService *serviceMock.ICustomerService)
		expectedStatusCode int
		useToken           bool
	}{
		{
			name: "SUCCESS: Delete customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("DeleteCustomer", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			},
			useToken:           true,
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Failed when update customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("DeleteCustomer", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("update error"))
			},
			useToken:           true,
			expectedStatusCode: 500,
		},
		{
			name: "ERROR: Unauthenticated",
			setup: func(customerService *serviceMock.ICustomerService) {
			},
			useToken:           false,
			expectedStatusCode: 401,
		},
	}
	validator := validatorExt.New()
	token := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: "valid-merchant-id",
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerService := serviceMock.NewICustomerService(t)
			tc.setup(customerService)
			controller := New(customerService, validator)
			chiRouteCtx := chi.NewRouteContext()
			chiRouteCtx.URLParams.Add("id", uuid)
			req := httptest.NewRequest(http.MethodDelete, "/url", nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouteCtx))
			if tc.useToken {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, token))
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.Delete)
			handler.ServeHTTP(rr, req)
			if rr.Code != 200 {
				fmt.Println(rr.Body.String())
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

		})
	}
}
