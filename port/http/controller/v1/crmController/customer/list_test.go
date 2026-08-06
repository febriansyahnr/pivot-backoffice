package customer_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/customer"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetCustomerList(t *testing.T) {
	service := serviceMocks.NewICustomerService(t)

	router := chi.NewRouter()
	router.Get("/customers", New(service, validatorExt.New()).GetCustomerList)

	customers := []customerModel.GeneralCustomerResponse{
		{
			UUID:        "customer-uuid-1",
			MerchantID:  "merchant-123",
			Email:       "customer1@example.com",
			PhoneNumber: "08123456789",
			FirstName:   "John",
			LastName:    "Doe",
			IsBlocked:   false,
			BlockReason: "",
		},
		{
			UUID:        "customer-uuid-2",
			MerchantID:  "merchant-123",
			Email:       "customer2@example.com",
			PhoneNumber: "08987654321",
			FirstName:   "Jane",
			LastName:    "Smith",
			IsBlocked:   false,
			BlockReason: "",
		},
	}

	meta := commonModel.Meta{
		Page:       1,
		PerPage:    10,
		TotalItems: 2,
		TotalPages: 1,
	}

	paginationResponse := &commonModel.PaginationResponse{
		Data: customers,
		Meta: meta,
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Missing merchant_id parameter",
			queryParams:    "",
			setupMock:      func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"merchant id is required","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "?merchant_id=merchant-123",
			setupMock: func() {
				service.On(
					"GetCustomerList", c.ValueCtxMockType(), "merchant-123", "", int64(1), int64(10),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Get customer list without filters",
			queryParams: "?merchant_id=merchant-123",
			setupMock: func() {
				service.On(
					"GetCustomerList", c.ValueCtxMockType(), "merchant-123", "", int64(1), int64(10),
				).Once().Return(paginationResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[{"uuid":"customer-uuid-1","merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","businessName":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","city":"","country":"","addressLine1":"","addressLine2":"","postalCode":"","state":"","isBlocked":false,"blockReason":"","metadata":null},{"uuid":"customer-uuid-2","merchantId":"merchant-123","email":"customer2@example.com","phoneNumber":"08987654321","firstName":"Jane","lastName":"Smith","businessName":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","city":"","country":"","addressLine1":"","addressLine2":"","postalCode":"","state":"","isBlocked":false,"blockReason":"","metadata":null}],"meta":{"page":1,"perPage":10,"totalItems":2,"totalPages":1}}}`,
		},
		{
			name:        "SUCCESS: Get customer list with phone number filter",
			queryParams: "?merchant_id=merchant-123&phone_number=08123456789",
			setupMock: func() {
				service.On(
					"GetCustomerList", c.ValueCtxMockType(), "merchant-123", "08123456789", int64(1), int64(10),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []customerModel.GeneralCustomerResponse{customers[0]},
					Meta: commonModel.Meta{Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[{"uuid":"customer-uuid-1","merchantId":"merchant-123","email":"customer1@example.com","phoneNumber":"08123456789","firstName":"John","lastName":"Doe","businessName":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","city":"","country":"","addressLine1":"","addressLine2":"","postalCode":"","state":"","isBlocked":false,"blockReason":"","metadata":null}],"meta":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}}`,
		},
		{
			name:        "SUCCESS: Get customer list with pagination",
			queryParams: "?merchant_id=merchant-123&page=2&per_page=5",
			setupMock: func() {
				service.On(
					"GetCustomerList", c.ValueCtxMockType(), "merchant-123", "", int64(2), int64(5),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []customerModel.GeneralCustomerResponse{},
					Meta: commonModel.Meta{Page: 2, PerPage: 5, TotalItems: 2, TotalPages: 1},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"data":[],"meta":{"page":2,"perPage":5,"totalItems":2,"totalPages":1}}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/customers"+test.queryParams, nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), fmt.Sprintf("expected: %s, actual: %s", test.wantRespBody, rec.Body.String()))
		})
	}
}
