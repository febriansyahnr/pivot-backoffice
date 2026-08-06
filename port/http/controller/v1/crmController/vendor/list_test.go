package vendor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	expectedSuccessWithNameFilter, _ := json.Marshal(map[string]interface{}{
		"code": "00",
		"data": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":                "123e4567-e89b-12d3-a456-426614174000",
					"merchantId":          "merchant-123",
					"name":                "Test Vendor",
					"beneficialOwner":     "John Doe",
					"businessCategory":    "E-Commerce",
					"avgMonthlyTpvAmount": "1000000",
					"bankName":            "Bank ABC",
					"bankCode":            "ABC",
					"accountNumber":       "1234567890",
					"accountName":         "Test Account",
					"status":              "ACTIVE",
					"createdAt":           "0001-01-01T00:00:00Z",
					"updatedAt":           "0001-01-01T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":       float64(1),
				"perPage":    float64(10),
				"totalItems": float64(1),
				"totalPages": float64(1),
			},
		},
		"message": "Success",
	})

	expectedSuccessWithMerchantIdFilter, _ := json.Marshal(map[string]interface{}{
		"code": "00",
		"data": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":                "123e4567-e89b-12d3-a456-426614174000",
					"merchantId":          "550e8400-e29b-41d4-a716-446655440000",
					"name":                "Test Vendor",
					"beneficialOwner":     "John Doe",
					"businessCategory":    "E-Commerce",
					"avgMonthlyTpvAmount": "1000000",
					"bankName":            "Bank ABC",
					"bankCode":            "ABC",
					"accountNumber":       "1234567890",
					"accountName":         "Test Account",
					"status":              "ACTIVE",
					"createdAt":           "0001-01-01T00:00:00Z",
					"updatedAt":           "0001-01-01T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":       float64(1),
				"perPage":    float64(10),
				"totalItems": float64(1),
				"totalPages": float64(1),
			},
		},
		"message": "Success",
	})

	expectedSuccessWithSorting, _ := json.Marshal(map[string]interface{}{
		"code": "00",
		"data": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":                "123e4567-e89b-12d3-a456-426614174000",
					"merchantId":          "merchant-123",
					"name":                "Test Vendor",
					"beneficialOwner":     "John Doe",
					"businessCategory":    "E-Commerce",
					"avgMonthlyTpvAmount": "1000000",
					"bankName":            "Bank ABC",
					"bankCode":            "ABC",
					"accountNumber":       "1234567890",
					"accountName":         "Test Account",
					"status":              "ACTIVE",
					"createdAt":           "0001-01-01T00:00:00Z",
					"updatedAt":           "0001-01-01T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":       float64(1),
				"perPage":    float64(10),
				"totalItems": float64(1),
				"totalPages": float64(1),
			},
		},
		"message": "Success",
	})

	testCases := []struct {
		name         string
		query        string
		setup        func(svc *mockSvc.IVendorService)
		expectedCode int
		expectedBody string
	}{
		{
			name:  "SUCCESS: List Vendors",
			query: "?name=Test&status=ACTIVE&page=1&perPage=10",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, &vendorModel.VendorQuery{
					Name:     "Test",
					Status:   "ACTIVE",
					Page:     1,
					PageSize: 10,
				}).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{
						{
							UUID:                "123e4567-e89b-12d3-a456-426614174000",
							MerchantID:          "merchant-123",
							Name:                "Test Vendor",
							BeneficialOwner:     "John Doe",
							BusinessCategory:    "E-Commerce",
							AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
							BankName:            "Bank ABC",
							BankCode:            "ABC",
							AccountNumber:       "1234567890",
							AccountName:         "Test Account",
							Status:              "ACTIVE",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccessWithNameFilter),
		},
		{
			name:  "SUCCESS: List Vendors with merchantId filter",
			query: "?merchantId=550e8400-e29b-41d4-a716-446655440000&page=1&perPage=10",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, &vendorModel.VendorQuery{
					MerchantID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Page:       1,
					PageSize:   10,
				}).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{
						{
							UUID:                "123e4567-e89b-12d3-a456-426614174000",
							MerchantID:          "550e8400-e29b-41d4-a716-446655440000",
							Name:                "Test Vendor",
							BeneficialOwner:     "John Doe",
							BusinessCategory:    "E-Commerce",
							AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
							BankName:            "Bank ABC",
							BankCode:            "ABC",
							AccountNumber:       "1234567890",
							AccountName:         "Test Account",
							Status:              "ACTIVE",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccessWithMerchantIdFilter),
		},
		{
			name:  "SUCCESS: List Vendors with sortBy and sort ASC",
			query: "?page=1&perPage=10&sortBy=createdAt&sort=ASC",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, &vendorModel.VendorQuery{
					Page:     1,
					PageSize: 10,
					SortBy:   "createdAt",
					Sort:     "ASC",
				}).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{
						{
							UUID:                "123e4567-e89b-12d3-a456-426614174000",
							MerchantID:          "merchant-123",
							Name:                "Test Vendor",
							BeneficialOwner:     "John Doe",
							BusinessCategory:    "E-Commerce",
							AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
							BankName:            "Bank ABC",
							BankCode:            "ABC",
							AccountNumber:       "1234567890",
							AccountName:         "Test Account",
							Status:              "ACTIVE",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccessWithSorting),
		},
		{
			name:  "SUCCESS: List Vendors with sortBy and sort DESC",
			query: "?page=1&perPage=10&sortBy=createdAt&sort=DESC",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, &vendorModel.VendorQuery{
					Page:     1,
					PageSize: 10,
					SortBy:   "createdAt",
					Sort:     "DESC",
				}).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{
						{
							UUID:                "123e4567-e89b-12d3-a456-426614174000",
							MerchantID:          "merchant-123",
							Name:                "Test Vendor",
							BeneficialOwner:     "John Doe",
							BusinessCategory:    "E-Commerce",
							AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
							BankName:            "Bank ABC",
							BankCode:            "ABC",
							AccountNumber:       "1234567890",
							AccountName:         "Test Account",
							Status:              "ACTIVE",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccessWithSorting),
		},
		{
			name:         "ERROR: Invalid MerchantId",
			query:        "?merchantId=invalid-uuid&page=1&perPage=10",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid merchant id",
				"error": {
					"type": "API_ERROR",
					"message": "invalid merchant id", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid Page Param - non numeric",
			query:        "?page=zero",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid Page Param - zero",
			query:        "?page=0",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid Page Param - negative",
			query:        "?page=-1",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid PerPage Param - negative",
			query:        "?perPage=-1",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid per page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid per page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid PerPage Param - zero",
			query:        "?perPage=0",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid per page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid per page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid PerPage Param - non numeric",
			query:        "?perPage=abc",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid per page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid per page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid StartDate format",
			query:        "?startDate=invalid-date",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format",
				"error": {
					"type": "API_ERROR",
					"message": "invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid EndDate format",
			query:        "?endDate=invalid-date",
			setup:        func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format",
				"error": {
					"type": "API_ERROR",
					"message": "invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:  "ERROR: Service Error",
			query: "?name=fail",
			setup: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, mock.Anything).
					Return(nil, errors.New("something went wrong"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "something went wrong",
				"error": {
					"type": "UNKNOWN",
					"message": "something went wrong", "recommendation": ""
				},
				"data": null
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIVendorService(t)
			validate := validator.New()

			tc.setup(svc)

			ctrl := New(svc, validate)

			req := httptest.NewRequest(http.MethodGet, "/card-funded-payout/vendors"+tc.query, nil)
			req = req.WithContext(ctx)

			req.URL.RawQuery = url.Values(req.URL.Query()).Encode()

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.List)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
		})
	}
}
