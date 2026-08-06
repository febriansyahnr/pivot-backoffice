package customerController

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestUpdate(t *testing.T) {
	uuid := uuid.New().String()
	testCases := []struct {
		name                 string
		setup                func(customerService *serviceMock.ICustomerService)
		request              []byte
		expectedStatusCode   int
		useToken             bool
		expectedResponseBody []byte
	}{
		{
			name: "SUCCESS: Update customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("UpdateCustomer", mock.Anything, mock.Anything).Return(nil, nil)
			},
			request: []byte(`
				{
					"firstName":"John",
					"lastName":"Doe",
					"merchantId":"123",
					"phoneNumber":"081234567890"
				}`),
			useToken:           true,
			expectedStatusCode: 200,
			expectedResponseBody: []byte(`
			{
				"code": "01",
				"message": "Updated",
				"data": {
					"uuid": "5d3e42e7-13b8-4f19-a8e7-459e3a183c00",
					"merchantId": "364a262f-8487-474d-904a-746b81c4b121",
					"phoneNumber": "081234567890",
					"firstName": "John",
					"lastName": "Doe",
				}
			}
		`),
		},
		{
			name: "ERROR: Failed when deconstruct payload",
			setup: func(customerService *serviceMock.ICustomerService) {
			},
			request: []byte(`
			{
			#
			`),
			useToken:           true,
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Failed when update customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("UpdateCustomer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("update error"))
			},
			request: []byte(`
			{
				"firstName":"John",
				"lastName":"Doe",
				"merchantId":"123",
				"phoneNumber":"081234567890"
			}`),
			useToken:           true,
			expectedStatusCode: 500,
		},
		{
			name: "ERROR: Unauthenticated",
			setup: func(customerService *serviceMock.ICustomerService) {
			},
			request: []byte(`
			{
				"firstName":"John",
				"lastName":"Doe",
				"merchantId":"123",
				"phoneNumber":"081234567890"
			}`),
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
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewBuffer(tc.request))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouteCtx))
			if tc.useToken {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, token))
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.Update)
			handler.ServeHTTP(rr, req)
			if rr.Code == 200 {
				var response CreateCustomerJSONResponse
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				fmt.Print(response)
				var expectedResponse CreateCustomerJSONResponse
				json.Unmarshal(tc.expectedResponseBody, &expectedResponse)
				assert.Equal(t, response.Data.PhoneNumber, expectedResponse.Data.PhoneNumber)
				assert.Equal(t, response.Data.FirstName, expectedResponse.Data.FirstName)
				assert.Equal(t, response.Data.LastName, expectedResponse.Data.LastName)
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

		})
	}
}
