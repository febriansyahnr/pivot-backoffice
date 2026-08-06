package customerController

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CreateCustomerJSONResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UUID        string `json:"uuid"`
		MerchantID  string `json:"merchantId"`
		PhoneNumber string `json:"phoneNumber"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
	} `json:"data"`
}

type CustomerErrorJSONResponse struct {
	Code   string `json:"code"`
	Errors string `json:"errors"`
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name                 string
		setup                func(customerService *serviceMock.ICustomerService)
		request              []byte
		expectedStatusCode   int
		useToken             bool
		expectedResponseBody []byte
	}{
		{
			name: "SUCCESS: Create customer and wallet",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("CreateCustomer", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{
					FirstName:   "John",
					LastName:    "Doe",
					MerchantID:  "364a262f-8487-474d-904a-746b81c4b121",
					PhoneNumber: "081234567890",
				}, nil)
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
					"message": "Created",
					"data": {
						"uuid": "5d3e42e7-13b8-4f19-a8e7-459e3a183c00",
						"merchantId": "364a262f-8487-474d-904a-746b81c4b121",
						"phoneNumber": "081234567890",
						"firstName": "John",
						"lastName": "Doe"
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
			expectedResponseBody: []byte(`
			{
				"code": "40",
			  "errors": "unexpected EOF"
			}`),
		},
		{
			name: "ERROR: Failed when create customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("CreateCustomer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Create error"))
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
			expectedResponseBody: []byte(`
			{
				"code": "41",
			  "errors": "missing token"
			}`),
		},
		{
			name: "ERROR: Invalid request",
			setup: func(customerService *serviceMock.ICustomerService) {
			},
			request: []byte(`
			{
				"firstName":"John",
				"lastName":"Doe",
				"merchantId":"123"
			}`),
			useToken:           true,
			expectedStatusCode: 400,
			expectedResponseBody: []byte(`
			{
				"code": "40",
			  "errors": "phoneNumber is required"
			}`),
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
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewBuffer(tc.request))
			if tc.useToken {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, token))
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.Create)
			handler.ServeHTTP(rr, req)
			if rr.Code == 201 {
				var response CreateCustomerJSONResponse
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				var expectedResponse CreateCustomerJSONResponse
				_ = json.Unmarshal(tc.expectedResponseBody, &expectedResponse)
				assert.Equal(t, response.Data.PhoneNumber, expectedResponse.Data.PhoneNumber)
				assert.Equal(t, response.Data.FirstName, expectedResponse.Data.FirstName)
				assert.Equal(t, response.Data.LastName, expectedResponse.Data.LastName)
			} else {
				var response CustomerErrorJSONResponse
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				var expectedResponse CustomerErrorJSONResponse
				_ = json.Unmarshal(tc.expectedResponseBody, &expectedResponse)
				print(tc.name, ": ", response.Errors)
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

		})
	}
}

func TestCreateWalletCustomer(t *testing.T) {
	testCases := []struct {
		name                 string
		setup                func(customerService *serviceMock.ICustomerService)
		request              []byte
		expectedStatusCode   int
		useMerchantId        bool
		expectedResponseBody []byte
	}{
		{
			name: "SUCCESS: Create customer and wallet",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("CreateCustomer", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{
					FirstName:   "John",
					LastName:    "Doe",
					MerchantID:  "364a262f-8487-474d-904a-746b81c4b121",
					PhoneNumber: "081234567890",
				}, nil)
			},
			request: []byte(`
				{
					"firstName":"John",
					"lastName":"Doe",
					"merchantId":"123",
					"phoneNumber":"081234567890"
				}`),
			useMerchantId:      true,
			expectedStatusCode: 200,
			expectedResponseBody: []byte(`
				{
					"code": "01",
					"message": "Created",
					"data": {
						"uuid": "5d3e42e7-13b8-4f19-a8e7-459e3a183c00",
						"merchantId": "364a262f-8487-474d-904a-746b81c4b121",
						"phoneNumber": "081234567890",
						"firstName": "John",
						"lastName": "Doe"
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
			useMerchantId:      true,
			expectedStatusCode: 400,
			expectedResponseBody: []byte(`
			{
				"code": "40",
			  "errors": "unexpected EOF"
			}`),
		},
		{
			name: "ERROR: Failed when create customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("CreateCustomer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Create error"))
			},
			request: []byte(`
			{
				"firstName":"John",
				"lastName":"Doe",
				"merchantId":"123",
				"phoneNumber":"081234567890"
			}`),
			useMerchantId:      true,
			expectedStatusCode: 500,
		},
		{
			name: "ERROR: Invalid request",
			setup: func(customerService *serviceMock.ICustomerService) {
			},
			request: []byte(`
			{
				"firstName":"John",
				"lastName":"Doe",
				"merchantId":"123"
			}`),
			useMerchantId:      true,
			expectedStatusCode: 400,
			expectedResponseBody: []byte(`
			{
				"code": "40",
			  "errors": "phoneNumber is required"
			}`),
		},
	}
	validator := validatorExt.New()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerService := serviceMock.NewICustomerService(t)
			tc.setup(customerService)
			controller := New(customerService, validator)
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewBuffer(tc.request))
			if tc.useMerchantId {
				req.Header.Add(constant.HeaderXMerchantId, uuid.NewString())
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.CreateWalletCustomer)
			handler.ServeHTTP(rr, req)
			if rr.Code == 201 {
				var response CreateCustomerJSONResponse
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				var expectedResponse CreateCustomerJSONResponse
				_ = json.Unmarshal(tc.expectedResponseBody, &expectedResponse)
				assert.Equal(t, response.Data.PhoneNumber, expectedResponse.Data.PhoneNumber)
				assert.Equal(t, response.Data.FirstName, expectedResponse.Data.FirstName)
				assert.Equal(t, response.Data.LastName, expectedResponse.Data.LastName)
			} else {
				var response CustomerErrorJSONResponse
				_ = json.Unmarshal(rr.Body.Bytes(), &response)
				var expectedResponse CustomerErrorJSONResponse
				_ = json.Unmarshal(tc.expectedResponseBody, &expectedResponse)
				print(tc.name, ": ", response.Errors)
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

		})
	}
}
