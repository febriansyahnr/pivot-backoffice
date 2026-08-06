package customer_test

import (
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

func TestGetCustomerByID(t *testing.T) {
	service := serviceMocks.NewICustomerService(t)

	router := chi.NewRouter()
	router.Get("/customers/{id}", New(service, validatorExt.New()).GetCustomerByID)

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
		customerID     string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Missing merchant_id parameter",
			customerID:     "customer-uuid-1",
			queryParams:    "",
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"merchant id is required","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: Missing customer ID in path",
			customerID:     "",
			queryParams:    "?merchant_id=merchant-123",
			setupMock:      func() {},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   "404 page not found\n",
		},
		{
			name:        "ERROR: Service error",
			customerID:  "customer-uuid-1",
			queryParams: "?merchant_id=merchant-123",
			setupMock: func() {
				service.On(
					"GetCustomerById", c.ValueCtxMockType(), "customer-uuid-1", "merchant-123",
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Get customer by ID",
			customerID:  "customer-uuid-1",
			queryParams: "?merchant_id=merchant-123",
			setupMock: func() {
				service.On(
					"GetCustomerById", c.ValueCtxMockType(), "customer-uuid-1", "merchant-123",
				).Once().Return(customer, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"customer-uuid-1","merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","businessName":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","city":"","country":"","addressLine1":"","addressLine2":"","postalCode":"","state":"","isBlocked":false,"blockReason":"","metadata":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			path := "/customers/"
			if test.customerID != "" {
				path += test.customerID
			}
			req := httptest.NewRequest(http.MethodGet, path+test.queryParams, nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if test.wantStatusCode == http.StatusNotFound {
				assert.Equal(t, test.wantRespBody, rec.Body.String())
			} else {
				assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
			}
		})
	}
}