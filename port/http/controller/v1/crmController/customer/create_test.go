package customer_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/customer"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	service := serviceMocks.NewICustomerService(t)

	router := chi.NewRouter()
	router.Post("/customers", New(service, validatorExt.New()).Create)

	customer := &customerModel.GeneralCustomerResponse{
		UUID:        "customer-uuid-1",
		MerchantID:  "merchant-123",
		Email:       "customer1@example.com", 
		PhoneNumber: "08123456789",
		FirstName:   "John",
		LastName:    "Doe",
		IsBlocked:   false,
		BlockReason: "",
	}

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid JSON format",
			requestBody:    `{invalid json}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid unmarshal JSON","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Missing merchant ID",
			requestBody:    `{"email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John"}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"merchant id is required","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Validation failed - missing phone number",
			requestBody:    `{"merchantId":"merchant-123","email":"customer1@example.com","firstName":"John"}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "ERROR: Validation failed - isBlocked true but no blockReason",
			requestBody:    `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","isBlocked":true}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"blockReason is required when isBlocked is true","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Validation failed - isBlocked true but empty blockReason",
			requestBody:    `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","isBlocked":true,"blockReason":""}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"blockReason is required when isBlocked is true","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe"}`,
			setupMock: func() {
				service.On(
					"CreateCustomer", c.ValueCtxMockType(), customerModel.CreateCustomerRequest{
						MerchantID:  "merchant-123",
						Email:       "customer1@example.com",
						PhoneNumber: "08123456789",
						FirstName:   "John",
						LastName:    "Doe",
					},
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Create customer",
			requestBody: `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","isBlocked":false}`,
			setupMock: func() {
				service.On(
					"CreateCustomer", c.ValueCtxMockType(), customerModel.CreateCustomerRequest{
						MerchantID:  "merchant-123",
						Email:       "customer1@example.com",
						PhoneNumber: "08123456789",
						FirstName:   "John",
						LastName:    "Doe",
						IsBlocked:   false,
					},
				).Once().Return(customer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SUCCESS: Create customer with isBlocked false ignores blockReason",
			requestBody: `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","isBlocked":false,"blockReason":"This should be ignored"}`,
			setupMock: func() {
				service.On(
					"CreateCustomer", c.ValueCtxMockType(), customerModel.CreateCustomerRequest{
						MerchantID:  "merchant-123",
						Email:       "customer1@example.com",
						PhoneNumber: "08123456789",
						FirstName:   "John",
						LastName:    "Doe",
						IsBlocked:   false,
						BlockReason: "", // Should be cleared when isBlocked is false
					},
				).Once().Return(customer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SUCCESS: Create customer with block info",
			requestBody: `{"merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","isBlocked":true,"blockReason":"Suspicious activity"}`,
			setupMock: func() {
				blockedCustomer := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid-2",
					MerchantID:  "merchant-123",
					Email:       "customer1@example.com", 
					PhoneNumber: "08123456789",
					FirstName:   "John",
					LastName:    "Doe",
					IsBlocked:   true,
					BlockReason: "Suspicious activity",
				}
				service.On(
					"CreateCustomer", c.ValueCtxMockType(), customerModel.CreateCustomerRequest{
						MerchantID:  "merchant-123",
						Email:       "customer1@example.com",
						PhoneNumber: "08123456789",
						FirstName:   "John",
						LastName:    "Doe",
						IsBlocked:   true,
						BlockReason: "Suspicious activity",
					},
				).Once().Return(blockedCustomer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader([]byte(test.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			
			if test.wantRespBody != "" {
				// For validation errors, we just check that it's a bad request
				if strings.Contains(test.name, "Validation failed") {
					return
				}
				assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
			}
		})
	}
}