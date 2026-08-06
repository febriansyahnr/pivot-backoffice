package charges

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestChargesController_GetChargeList(t *testing.T) {
	mockUnifiedPaymentService := serviceMocks.NewIUnifiedPaymentService(t)
	mockMerchantService := serviceMocks.NewIMerchantService(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	controller := &ChargesController{
		unifiedPaymentService: mockUnifiedPaymentService,
		merchantService:       mockMerchantService,
		logger:                mockLogger,
	}

	testCases := []struct {
		name           string
		setupRequest   func() *http.Request
		setupContext   func() context.Context
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "SUCCESS: Get charge list",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges?page=1&perPage=10", nil)
				return req
			},
			setupContext: func() context.Context {
				user := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, user)
				return ctx
			},
			setupMocks: func() {
				expectedResult := &commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:     "charge-123",
							Status: "SUCCESS",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}

				mockUnifiedPaymentService.On(
					"GetChargeList",
					mock.Anything,
					mock.MatchedBy(func(req *unifiedPaymentModel.FilterChargeRequest) bool {
						return req.MerchantID == "merchant-123" && req.Page == 1 && req.PerPage == 10
					}),
				).Return(expectedResult, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: Get charge list with sub-merchant",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges?subMerchantId=sub-merchant-123", nil)
				return req
			},
			setupContext: func() context.Context {
				user := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, user)
				return ctx
			},
			setupMocks: func() {
				mockMerchantService.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					"merchant-123",
					"sub-merchant-123",
				).Return(nil)

				expectedResult := &commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{
						{
							ID:     "charge-456",
							Status: "PENDING",
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					},
				}

				mockUnifiedPaymentService.On(
					"GetChargeList",
					mock.Anything,
					mock.MatchedBy(func(req *unifiedPaymentModel.FilterChargeRequest) bool {
						return req.MerchantID == "sub-merchant-123"
					}),
				).Return(expectedResult, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ERROR: User not found in context",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges", nil)
				return req
			},
			setupContext: func() context.Context {
				return context.Background() // No user in context
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Invalid sub-merchant validation",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges?subMerchantId=invalid-sub-merchant", nil)
				return req
			},
			setupContext: func() context.Context {
				user := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, user)
				return ctx
			},
			setupMocks: func() {
				mockMerchantService.On(
					"ValidateSubMerchantParent",
					mock.Anything,
					"merchant-123",
					"invalid-sub-merchant",
				).Return(pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: Invalid date range value",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges?startDate=2025-01-01T16:53:00.000Z&endDate=2025-01-31T16:52:59.999Z", nil)
				return req
			},
			setupContext: func() context.Context {
				user := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, user)
				return ctx
			},
			setupMocks:     func() { /* No Body */ },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: Service error",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/charges", nil)
				return req
			},
			setupContext: func() context.Context {
				user := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, user)
				return ctx
			},
			setupMocks: func() {
				mockUnifiedPaymentService.On(
					"GetChargeList",
					mock.Anything,
					mock.Anything,
				).Return(nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			req := tc.setupRequest()
			ctx := tc.setupContext()
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router := chi.NewRouter()
			router.Get("/api/v1/charges", controller.GetChargeList)

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Reset mocks for next test
			mockUnifiedPaymentService.ExpectedCalls = nil
			mockMerchantService.ExpectedCalls = nil
		})
	}
}

func TestChargesController_parseChargeFilterParam(t *testing.T) {
	controller := &ChargesController{}

	testCases := []struct {
		name        string
		queryParams map[string]string
		expected    unifiedPaymentModel.FilterChargeRequest
		expectError bool
	}{
		{
			name:        "Default values",
			queryParams: map[string]string{},
			expected: unifiedPaymentModel.FilterChargeRequest{
				Page:    1,
				PerPage: 10,
				Sort:    "ASC",
				SortBy:  "createdAt",
			},
			expectError: false,
		},
		{
			name: "Custom pagination",
			queryParams: map[string]string{
				"page":    "2",
				"perPage": "20",
			},
			expected: unifiedPaymentModel.FilterChargeRequest{
				Page:    2,
				PerPage: 20,
				Sort:    "ASC",
				SortBy:  "createdAt",
			},
			expectError: false,
		},
		{
			name: "With filters",
			queryParams: map[string]string{
				"id":                "charge-123",
				"status":            "SUCCESS",
				"clientReferenceId": "ref-123",
				"paymentSessionId":  "session-123",
				"sort":              "DESC",
				"sortBy":            "updatedAt",
			},
			expected: unifiedPaymentModel.FilterChargeRequest{
				Page:              1,
				PerPage:           10,
				Sort:              "DESC",
				SortBy:            "updatedAt",
				UUID:              "charge-123",
				Status:            "SUCCESS",
				ClientReferenceID: "ref-123",
				PaymentSessionID:  "session-123",
			},
			expectError: false,
		},
		{
			name: "Invalid page format",
			queryParams: map[string]string{
				"page": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid perPage format",
			queryParams: map[string]string{
				"perPage": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid date format",
			queryParams: map[string]string{
				"startDate": "invalid-date",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build URL with query parameters
			u, _ := url.Parse("/api/v1/charges")
			q := u.Query()
			for key, value := range tc.queryParams {
				q.Set(key, value)
			}
			u.RawQuery = q.Encode()

			req := httptest.NewRequest("GET", u.String(), nil)

			result, err := controller.parseChargeFilterParam(req)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.Page, result.Page)
				assert.Equal(t, tc.expected.PerPage, result.PerPage)
				assert.Equal(t, tc.expected.Sort, result.Sort)
				assert.Equal(t, tc.expected.SortBy, result.SortBy)
				assert.Equal(t, tc.expected.UUID, result.UUID)
				assert.Equal(t, tc.expected.Status, result.Status)
				assert.Equal(t, tc.expected.ClientReferenceID, result.ClientReferenceID)
				assert.Equal(t, tc.expected.PaymentSessionID, result.PaymentSessionID)
			}
		})
	}
}
