package cardFundedPayoutController

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVendorList(t *testing.T) {
	merchantID := "550e8400-e29b-41d4-a716-446655440000"

	testCases := []struct {
		name         string
		query        string
		setupContext func() context.Context
		setupMock    func(svc *mockSvc.IVendorService)
		expectedCode int
	}{
		{
			name:  "SUCCESS: Get vendor list with pagination",
			query: "?page=1&perPage=10",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{
						{
							UUID:                "123e4567-e89b-12d3-a456-426614174000",
							MerchantID:          merchantID,
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
		},
		{
			name:  "SUCCESS: Get vendor list with default pagination",
			query: "",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "SUCCESS: Get vendor list with name and status filters",
			query: "?name=Test&status=ACTIVE",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, mock.MatchedBy(func(q *vendorModel.VendorQuery) bool {
					return q.Name == "Test" && q.Status == "ACTIVE"
				})).Return(&commonModel.PaginationResponse{
					Data: []*vendorModel.VendorResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "ERROR: User not in context",
			query: "?page=1&perPage=10",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:  "ERROR: Invalid merchantId in token",
			query: "?page=1&perPage=10",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: "invalid-uuid",
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Invalid page param - non numeric",
			query: "?page=abc",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Invalid page param - zero",
			query: "?page=0",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Invalid perPage param - negative",
			query: "?perPage=-1",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Invalid perPage param - non numeric",
			query: "?perPage=abc",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Invalid perPage param - zero",
			query: "?perPage=0",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "ERROR: Service error",
			query: "?page=1&perPage=10",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vendorSvc := mockSvc.NewIVendorService(t)
			tc.setupMock(vendorSvc)

			ctrl := &handler{
				vendorService: vendorSvc,
			}

			req := httptest.NewRequest(http.MethodGet, "/card-funded-payouts/vendors"+tc.query, nil)
			req = req.WithContext(tc.setupContext())
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.GetVendorList)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}

func TestGetVendorDetail(t *testing.T) {
	merchantID := "550e8400-e29b-41d4-a716-446655440000"
	vendorUUID := "123e4567-e89b-12d3-a456-426614174000"

	existingVendor := &vendorModel.Vendor{
		UUID:                vendorUUID,
		MerchantID:          merchantID,
		Name:                "Test Vendor",
		BeneficialOwner:     "John Doe",
		BusinessCategory:    "E-Commerce",
		AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
		BankName:            "Bank ABC",
		BankCode:            "ABC",
		AccountNumber:       "1234567890",
		AccountName:         "Test Account",
		Status:              "ACTIVE",
	}

	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code":    "00",
		"message": "OK",
		"data": map[string]interface{}{
			"uuid":                vendorUUID,
			"merchantId":          merchantID,
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
	})

	testCases := []struct {
		name         string
		vendorID     string
		setupContext func() context.Context
		setupMock    func(svc *mockSvc.IVendorService)
		expectedCode int
		expectedBody string
	}{
		{
			name:     "SUCCESS: Get vendor detail",
			vendorID: vendorUUID,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("Detail", mock.Anything, vendorUUID).Return(existingVendor, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name:     "ERROR: User not in context",
			vendorID: vendorUUID,
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:     "ERROR: Invalid vendor UUID",
			vendorID: "invalid-uuid",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock:    func(svc *mockSvc.IVendorService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:     "ERROR: Vendor belongs to different merchant",
			vendorID: vendorUUID,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: "660e8400-e29b-41d4-a716-446655440001", // different merchant
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("Detail", mock.Anything, vendorUUID).Return(existingVendor, nil)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:     "ERROR: Service error",
			vendorID: vendorUUID,
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantID,
				})
			},
			setupMock: func(svc *mockSvc.IVendorService) {
				svc.On("Detail", mock.Anything, vendorUUID).Return(nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vendorSvc := mockSvc.NewIVendorService(t)
			tc.setupMock(vendorSvc)

			ctrl := &handler{
				vendorService: vendorSvc,
			}

			req := httptest.NewRequest(http.MethodGet, "/card-funded-payouts/vendors/"+tc.vendorID, nil)
			req = req.WithContext(tc.setupContext())

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.vendorID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.GetVendorDetail)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedBody != "" && tc.expectedCode == http.StatusOK {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}
