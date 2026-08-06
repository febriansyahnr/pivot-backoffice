package customer_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/customer"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	service := serviceMocks.NewICustomerService(t)

	router := chi.NewRouter()
	router.Put("/customers/{id}", New(service, validatorExt.New()).Update)

	customer := &customerModel.GeneralCustomerResponse{
		UUID:        "customer-uuid-1",
		MerchantID:  "merchant-123",
		Email:       "updated@example.com", 
		PhoneNumber: "08987654321",
		FirstName:   "Jane",
		LastName:    "Smith",
		IsBlocked:   true,
		BlockReason: "Fraud detected",
	}

	tests := []struct {
		name           string
		customerID     string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Missing customer ID in path",
			customerID:     "",
			requestBody:    `{"merchantId":"merchant-123","email":"updated@example.com"}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   "404 page not found\n",
		},
		{
			name:           "ERROR: Invalid JSON format",
			customerID:     "customer-uuid-1",
			requestBody:    `{invalid json}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid unmarshal JSON","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Missing merchant ID",
			customerID:     "customer-uuid-1",
			requestBody:    `{"email":"updated@example.com"}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"merchant id is required","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Validation failed - isBlocked true but no blockReason",
			customerID:     "customer-uuid-1",
			requestBody:    `{"merchantId":"merchant-123","isBlocked":true}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"blockReason is required when isBlocked is true","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Validation failed - isBlocked true but empty blockReason",
			customerID:     "customer-uuid-1",
			requestBody:    `{"merchantId":"merchant-123","isBlocked":true,"blockReason":""}`,
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"blockReason is required when isBlocked is true","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			customerID:  "customer-uuid-1",
			requestBody: `{"merchantId":"merchant-123","email":"updated@example.com"}`,
			setupMock: func() {
				emailPtr := "updated@example.com"
				service.On(
					"UpdateCustomer", c.ValueCtxMockType(), customerModel.UpdateCustomerRequest{
						UUID:       "customer-uuid-1",
						MerchantID: "merchant-123",
						Email:      &emailPtr,
					},
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Update customer email",
			customerID:  "customer-uuid-1",
			requestBody: `{"merchantId":"merchant-123","email":"updated@example.com"}`,
			setupMock: func() {
				emailPtr := "updated@example.com"
				service.On(
					"UpdateCustomer", c.ValueCtxMockType(), customerModel.UpdateCustomerRequest{
						UUID:       "customer-uuid-1",
						MerchantID: "merchant-123",
						Email:      &emailPtr,
					},
				).Once().Return(customer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SUCCESS: Update customer block status",
			customerID:  "customer-uuid-1",
			requestBody: `{"merchantId":"merchant-123","isBlocked":true,"blockReason":"Fraud detected"}`,
			setupMock: func() {
				isBlockedPtr := true
				blockReasonPtr := "Fraud detected"
				service.On(
					"UpdateCustomer", c.ValueCtxMockType(), customerModel.UpdateCustomerRequest{
						UUID:        "customer-uuid-1",
						MerchantID:  "merchant-123",
						IsBlocked:   &isBlockedPtr,
						BlockReason: &blockReasonPtr,
					},
				).Once().Return(customer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SUCCESS: Unblock customer - blockReason cleared when isBlocked false",
			customerID:  "customer-uuid-1",
			requestBody: `{"merchantId":"merchant-123","isBlocked":false,"blockReason":"Should be ignored"}`,
			setupMock: func() {
				isBlockedPtr := false
				blockReasonPtr := "" // Should be cleared automatically
				unblockedCustomer := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid-1",
					MerchantID:  "merchant-123",
					Email:       "updated@example.com", 
					PhoneNumber: "08987654321",
					FirstName:   "Jane",
					LastName:    "Smith",
					IsBlocked:   false,
					BlockReason: "",
				}
				service.On(
					"UpdateCustomer", c.ValueCtxMockType(), customerModel.UpdateCustomerRequest{
						UUID:        "customer-uuid-1",
						MerchantID:  "merchant-123",
						IsBlocked:   &isBlockedPtr,
						BlockReason: &blockReasonPtr,
					},
				).Once().Return(unblockedCustomer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "SUCCESS: Unblock customer",
			customerID:  "customer-uuid-1",
			requestBody: `{"merchantId":"merchant-123","isBlocked":false,"blockReason":""}`,
			setupMock: func() {
				isBlockedPtr := false
				blockReasonPtr := ""
				unblockedCustomer := &customerModel.GeneralCustomerResponse{
					UUID:        "customer-uuid-1",
					MerchantID:  "merchant-123",
					Email:       "updated@example.com", 
					PhoneNumber: "08987654321",
					FirstName:   "Jane",
					LastName:    "Smith",
					IsBlocked:   false,
					BlockReason: "",
				}
				service.On(
					"UpdateCustomer", c.ValueCtxMockType(), customerModel.UpdateCustomerRequest{
						UUID:        "customer-uuid-1",
						MerchantID:  "merchant-123",
						IsBlocked:   &isBlockedPtr,
						BlockReason: &blockReasonPtr,
					},
				).Once().Return(unblockedCustomer, nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			path := "/customers/"
			if test.customerID != "" {
				path += test.customerID
			}
			req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader([]byte(test.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			
			if test.wantRespBody != "" {
				if test.wantStatusCode == http.StatusNotFound {
					assert.Equal(t, test.wantRespBody, rec.Body.String())
				} else {
					assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
				}
			}
		})
	}
}