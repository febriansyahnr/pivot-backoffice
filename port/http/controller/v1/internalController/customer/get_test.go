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
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetById(t *testing.T) {
	validCustomerID := uuid.New().String()
	validMerchantID := uuid.New().String()
	testCases := []struct {
		name               string
		setup              func(customerService *serviceMock.ICustomerService)
		expectedStatusCode int
		customerID         string
		merchantID         string
	}{
		{
			name: "SUCCESS: Get customer by ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("FindCustomerByID", mock.Anything, validCustomerID).Return(nil, nil)
			},
			customerID:         validCustomerID,
			merchantID:         validMerchantID,
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Invalid customer ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			customerID:         "invalid-uuid",
			merchantID:         validMerchantID,
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Empty customer ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			customerID:         "",
			merchantID:         validMerchantID,
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Invalid merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// Service call still happens since merchant_id is not validated in GetById
				customerService.On("FindCustomerByID", mock.Anything, validCustomerID).Return(nil, nil)
			},
			customerID:         validCustomerID,
			merchantID:         "invalid-uuid",
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Empty merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// Service call still happens since merchant_id is not validated in GetById
				customerService.On("FindCustomerByID", mock.Anything, validCustomerID).Return(nil, nil)
			},
			customerID:         validCustomerID,
			merchantID:         "",
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Failed when get customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("FindCustomerByID", mock.Anything, validCustomerID).Return(nil, fmt.Errorf("service error"))
			},
			customerID:         validCustomerID,
			merchantID:         validMerchantID,
			expectedStatusCode: 500,
		},
	}
	validator := validatorExt.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerService := serviceMock.NewICustomerService(t)
			tc.setup(customerService)
			controller := New(customerService, validator)
			chiRouteCtx := chi.NewRouteContext()
			chiRouteCtx.URLParams.Add("id", tc.customerID)

			// Build URL with query parameters
			url := "/customers/" + tc.customerID + "?merchant_id=" + tc.merchantID
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouteCtx))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.GetById)
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.expectedStatusCode {
				t.Logf("Response body: %s", rr.Body.String())
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			customerService.AssertExpectations(t)
		})
	}
}

func TestGetList(t *testing.T) {
	validMerchantID := uuid.New().String()
	testCases := []struct {
		name               string
		setup              func(customerService *serviceMock.ICustomerService)
		expectedStatusCode int
		merchantID         string
		phoneNumber        string
	}{
		{
			name: "SUCCESS: Get customer list",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerList", mock.Anything, validMerchantID, "", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{}, nil)
			},
			merchantID:         validMerchantID,
			phoneNumber:        "",
			expectedStatusCode: 200,
		},
		{
			name: "SUCCESS: Get customer list with phone number",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerList", mock.Anything, validMerchantID, "081234567890", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{}, nil)
			},
			merchantID:         validMerchantID,
			phoneNumber:        "081234567890",
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Invalid merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			merchantID:         "invalid-uuid",
			phoneNumber:        "",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Empty merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			merchantID:         "",
			phoneNumber:        "",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Invalid phone number",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			merchantID:         validMerchantID,
			phoneNumber:        "invalid-phone",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Failed when get customer",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerList", mock.Anything, validMerchantID, "", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("service error"))
			},
			merchantID:         validMerchantID,
			phoneNumber:        "",
			expectedStatusCode: 500,
		},
	}
	validator := validatorExt.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerService := serviceMock.NewICustomerService(t)
			tc.setup(customerService)
			controller := New(customerService, validator)

			// Build URL with query parameters
			url := "/customers?merchant_id=" + tc.merchantID
			if tc.phoneNumber != "" {
				url += "&phoneNumber=" + tc.phoneNumber
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.GetList)
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.expectedStatusCode {
				t.Logf("Response body: %s", rr.Body.String())
			}
			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			customerService.AssertExpectations(t)
		})
	}
}

func TestGetByPhoneNumber(t *testing.T) {
	validMerchantID := uuid.New().String()
	testCases := []struct {
		name               string
		setup              func(customerService *serviceMock.ICustomerService)
		expectedStatusCode int
		phoneNumber        string
		merchantID         string
	}{
		{
			name: "SUCCESS: Get customer by phone number",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerByPhoneNumber", mock.Anything, "081234567890", validMerchantID).Return(nil, nil)
			},
			phoneNumber:        "081234567890",
			merchantID:         validMerchantID,
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Invalid phone number - empty",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			phoneNumber:        "",
			merchantID:         validMerchantID,
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Invalid merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			phoneNumber:        "081234567890",
			merchantID:         "invalid-uuid",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Empty merchant ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			phoneNumber:        "081234567890",
			merchantID:         "",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Service error",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerByPhoneNumber", mock.Anything, "081234567890", validMerchantID).Return(nil, fmt.Errorf("service error"))
			},
			phoneNumber:        "081234567890",
			merchantID:         validMerchantID,
			expectedStatusCode: 500,
		},
	}

	validator := validatorExt.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerService := serviceMock.NewICustomerService(t)
			tc.setup(customerService)
			controller := New(customerService, validator)

			chiRouteCtx := chi.NewRouteContext()
			chiRouteCtx.URLParams.Add("phoneNumber", tc.phoneNumber)

			// Build URL with query parameters
			url := "/customers/phone/" + tc.phoneNumber + "?merchant_id=" + tc.merchantID
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouteCtx))
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantIDKey, tc.merchantID))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetByPhoneNumber)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Logf("Response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			customerService.AssertExpectations(t)
		})
	}
}

func TestGetByIDForUnifiedPayment(t *testing.T) {
	validUUID := uuid.New().String()

	testCases := []struct {
		name               string
		setup              func(customerService *serviceMock.ICustomerService)
		expectedStatusCode int
		useToken           bool
		customerID         string
	}{
		{
			name: "SUCCESS: Get customer by ID for unified payment",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerByIDForUnifiedPayment", mock.Anything, validUUID, "valid-merchant-id").Return(nil, nil)
			},
			useToken:           true,
			customerID:         validUUID,
			expectedStatusCode: 200,
		},
		{
			name: "ERROR: Invalid customer ID - not a UUID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			useToken:           true,
			customerID:         "invalid-uuid",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Empty customer ID",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to validation failure
			},
			useToken:           true,
			customerID:         "",
			expectedStatusCode: 400,
		},
		{
			name: "ERROR: Service error",
			setup: func(customerService *serviceMock.ICustomerService) {
				customerService.On("GetCustomerByIDForUnifiedPayment", mock.Anything, validUUID, "valid-merchant-id").Return(nil, fmt.Errorf("service error"))
			},
			useToken:           true,
			customerID:         validUUID,
			expectedStatusCode: 500,
		},
		{
			name: "ERROR: Unauthorized - no merchant token",
			setup: func(customerService *serviceMock.ICustomerService) {
				// No service call expected due to auth failure
			},
			useToken:           false,
			customerID:         validUUID,
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
			chiRouteCtx.URLParams.Add("id", tc.customerID)
			req := httptest.NewRequest(http.MethodGet, "/customers/"+tc.customerID+"/unified", nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouteCtx))

			if tc.useToken {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, token))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetByIDForUnifiedPayment)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Logf("Response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			customerService.AssertExpectations(t)
		})
	}
}
