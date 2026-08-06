package payment_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"
	"github.com/paper-indonesia/pdk/go/monitoring"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFilterStaticVaList(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		queryParams    string
		hasUserContext bool
		setupMock      func(*serviceMocks.IPaymentService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			queryParams:    "?page=1&perPage=10",
			hasUserContext: false,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Invalid page parameter",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=invalid&perPage=10",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Invalid perPage parameter",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=invalid",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Invalid startDate format",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=10&startDate=invalid-date",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Invalid endDate format",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=10&endDate=invalid-date",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "SUCCESS: Valid request with query filter",
			expectedStatus: http.StatusOK,
			queryParams:    "?page=1&perPage=10&query=VAname234&status=ACTIVE",
			hasUserContext: true,
			setupMock: func(svc *serviceMocks.IPaymentService) {
				expectedResponse := &commonModel.PaginationResponse{
					Data: []paymentModel.StaticVaListResponse{
						{
							UUID:        "payment-123",
							ReferenceID: "VAname234",
							VaNumber:    "1234567890987654",
							VaBank:      "BCA",
							Status:      "ACTIVE",
						},
					},
					Meta: commonModel.Meta{Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1},
				}
				svc.On("FilterStaticVaList", mock.Anything, mock.AnythingOfType("paymentModel.StaticVaFilterRequest")).
					Return(expectedResponse, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := serviceMocks.NewIPaymentService(t)
			tc.setupMock(paymentService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-va"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			controller.FilterStaticVaList(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetStaticVaDetail(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		hasUserContext bool
		setupMock      func(*serviceMocks.IPaymentService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			hasUserContext: false,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "SUCCESS: Valid request",
			expectedStatus: http.StatusOK,
			paymentID:      "payment-123",
			hasUserContext: true,
			setupMock: func(svc *serviceMocks.IPaymentService) {
				expectedResponse := &paymentModel.StaticVaDetailResponse{
					UUID:        "payment-123",
					ReferenceID: "VAname234",
					VaNumber:    "1234567890987654",
					VaBank:      "BCA",
					Status:      "ACTIVE",
				}
				svc.On("GetStaticVaDetail", mock.Anything, mock.AnythingOfType("paymentModel.StaticVaDetailRequest")).
					Return(expectedResponse, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := serviceMocks.NewIPaymentService(t)
			tc.setupMock(paymentService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-va/"+tc.paymentID, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			// Set URL param for paymentId
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("paymentId", tc.paymentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			controller.GetStaticVaDetail(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetStaticVaTransactions(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		queryParams    string
		hasUserContext bool
		setupMock      func(*serviceMocks.IPaymentService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=10",
			hasUserContext: false,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			queryParams:    "?page=1&perPage=10",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "ERROR: Invalid page parameter",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			queryParams:    "?page=invalid&perPage=10",
			hasUserContext: true,
			setupMock:      func(svc *serviceMocks.IPaymentService) {},
		},
		{
			name:           "SUCCESS: Valid request",
			expectedStatus: http.StatusOK,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=10&status=SUCCESS",
			hasUserContext: true,
			setupMock: func(svc *serviceMocks.IPaymentService) {
				expectedResponse := &commonModel.PaginationResponse{
					Data: []paymentModel.StaticVaTransactionItem{
						{
							UUID:        "tx-123",
							ReferenceID: "QA202501151023",
							AmountValue: "10000",
							Status:      "SUCCESS",
						},
					},
					Meta: commonModel.Meta{Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1},
				}
				svc.On("GetStaticVaTransactions", mock.Anything, mock.AnythingOfType("paymentModel.StaticVaTransactionFilterRequest")).
					Return(expectedResponse, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := serviceMocks.NewIPaymentService(t)
			tc.setupMock(paymentService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-va/"+tc.paymentID+"/transactions"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			// Set URL param for paymentId
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("paymentId", tc.paymentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			controller.GetStaticVaTransactions(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestDeactivateStaticVa(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		requestBody    string
		hasUserContext bool
		hasPinHeader   bool
		setupMock      func(*serviceMocks.IPaymentService, *serviceMocks.IUserService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: false,
			hasPinHeader:   true,
			setupMock:      func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {},
		},
		{
			name:           "ERROR: Missing PIN header",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasPinHeader:   false,
			setupMock:      func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {},
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasPinHeader:   true,
			setupMock: func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {
				uSvc.On("CheckCurrentPin", mock.Anything, "user-123", "123456").Return(nil)
			},
		},
		{
			name:           "ERROR: Invalid JSON body",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `invalid json`,
			hasUserContext: true,
			hasPinHeader:   true,
			setupMock: func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {
				uSvc.On("CheckCurrentPin", mock.Anything, "user-123", "123456").Return(nil)
			},
		},
		{
			name:           "ERROR: Invalid status value",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INVALID_STATUS"}`,
			hasUserContext: true,
			hasPinHeader:   true,
			setupMock: func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {
				uSvc.On("CheckCurrentPin", mock.Anything, "user-123", "123456").Return(nil)
			},
		},
		{
			name:           "SUCCESS: Valid deactivation request",
			expectedStatus: http.StatusOK,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasPinHeader:   true,
			setupMock: func(pSvc *serviceMocks.IPaymentService, uSvc *serviceMocks.IUserService) {
				uSvc.On("CheckCurrentPin", mock.Anything, "user-123", "123456").Return(nil)
				pSvc.On("DeactivateStaticVa", mock.Anything, "payment-123", "merchant-123",
					mock.AnythingOfType("paymentModel.StaticVaUpdateStatusRequest")).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := serviceMocks.NewIPaymentService(t)
			userService := serviceMocks.NewIUserService(t)
			tc.setupMock(paymentService, userService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
				WithUserService(userService),
			)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/payments/static-va/"+tc.paymentID+"/deactivate",
				strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					UUID:       "user-123",
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			if tc.hasPinHeader {
				// Base64 encoded "123456"
				req.Header.Set(constant.HeaderXRequestPIN, "MTIzNDU2")
			}

			// Set URL param for paymentId
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("paymentId", tc.paymentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			controller.DeactivateStaticVa(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetVARangeList(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		queryParams    string
		hasUserContext bool
		setupMocks     func(*serviceMocks.IMerchantService, *serviceMocks.IPaymentMethodService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			queryParams:    "?status=ACTIVE",
			hasUserContext: false,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				// No mocks needed for this test case
			},
		},
		{
			name:           "ERROR: Merchant not found",
			expectedStatus: http.StatusUnprocessableEntity,
			queryParams:    "?status=ACTIVE",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(nil, nil)
			},
		},
		{
			name:           "ERROR: Database error when getting merchant",
			expectedStatus: http.StatusInternalServerError,
			queryParams:    "?status=ACTIVE",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(nil, assert.AnError)
			},
		},
		{
			name:           "SUCCESS: Get VA range list without parent merchant",
			expectedStatus: http.StatusOK,
			queryParams:    "?status=ACTIVE",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("GetStaticVAPaymentMethodByMerchant", mock.Anything,
					mock.MatchedBy(func(req *paymentModel.GetPaymentMethodFilterRequest) bool {
						return req.MerchantID == "merchant-123" && req.Status == "ACTIVE"
					})).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
			},
		},
		{
			name:           "SUCCESS: Get VA range list with parent merchant",
			expectedStatus: http.StatusOK,
			queryParams:    "?status=ACTIVE",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID:      "merchant-123",
					MID:       sql.NullString{String: "MID123", Valid: true},
					ParentID:  sql.NullString{String: "parent-merchant-123", Valid: true},
					KYCStatus: sql.NullString{String: constant.KYCStatusNotRequired, Valid: true},
				}
				parentMerchant := &merchantModel.Merchant{
					UUID: "parent-merchant-123",
					MID:  sql.NullString{String: "PARENT_MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(parentMerchant, nil)

				paymentMethodSvc.On("GetStaticVAPaymentMethodByMerchant", mock.Anything,
					mock.MatchedBy(func(req *paymentModel.GetPaymentMethodFilterRequest) bool {
						return req.MerchantID == "parent-merchant-123" && req.Status == "ACTIVE"
					})).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
			},
		},
		{
			name:           "SUCCESS: Get VA range list without status filter",
			expectedStatus: http.StatusOK,
			queryParams:    "",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("GetStaticVAPaymentMethodByMerchant", mock.Anything,
					mock.MatchedBy(func(req *paymentModel.GetPaymentMethodFilterRequest) bool {
						return req.MerchantID == "merchant-123" && req.Status == ""
					})).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
			},
		},
		{
			name:           "ERROR: Payment method service error",
			expectedStatus: http.StatusInternalServerError,
			queryParams:    "?status=ACTIVE",
			hasUserContext: true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("GetStaticVAPaymentMethodByMerchant", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantService := &serviceMocks.IMerchantService{}
			paymentMethodService := &serviceMocks.IPaymentMethodService{}
			tc.setupMocks(merchantService, paymentMethodService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithMerchantService(merchantService),
				WithPaymentMethodService(paymentMethodService),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-va/range"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			controller.GetVARangeList(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			merchantService.AssertExpectations(t)
			paymentMethodService.AssertExpectations(t)
		})
	}
}

func TestUpdateVARange(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name            string
		expectedStatus  int
		paymentMethodID string
		requestBody     string
		hasUserContext  bool
		setupMocks      func(*serviceMocks.IMerchantService, *serviceMocks.IPaymentMethodService)
	}{
		{
			name:            "ERROR: Missing user context",
			expectedStatus:  http.StatusUnauthorized,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  false,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				// No mocks needed for this test case
			},
		},
		{
			name:            "ERROR: Invalid payment method ID",
			expectedStatus:  http.StatusBadRequest,
			paymentMethodID: "invalid-uuid",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				// No mocks needed for this test case
			},
		},
		{
			name:            "ERROR: Invalid JSON body",
			expectedStatus:  http.StatusBadRequest,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": }`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				// No mocks needed for this test case
			},
		},
		{
			name:            "SUCCESS: Empty ranges should work since no validation tags",
			expectedStatus:  http.StatusOK,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "", "start": "", "end": ""}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:            "ERROR: Merchant not found",
			expectedStatus:  http.StatusUnprocessableEntity,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(nil, nil)
			},
		},
		{
			name:            "ERROR: Database error when getting merchant",
			expectedStatus:  http.StatusInternalServerError,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(nil, assert.AnError)
			},
		},
		{
			name:            "SUCCESS: Update VA range with close range only",
			expectedStatus:  http.StatusOK,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything,
					mock.MatchedBy(func(req *paymentMethodModel.SetupPaymentMethodConfigRequest) bool {
						return req.MerchantID == "merchant-123" &&
							req.PaymentMethodID == "550e8400-e29b-41d4-a716-446655440000" &&
							len(req.PartnerConfig.VirtualAccount.Items) == 1
					})).Return(nil)
			},
		},
		{
			name:            "SUCCESS: Update VA range with open range only",
			expectedStatus:  http.StatusOK,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"openRange": {"binPrefix": "54321", "start": "100", "end": "500"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything,
					mock.MatchedBy(func(req *paymentMethodModel.SetupPaymentMethodConfigRequest) bool {
						return req.MerchantID == "merchant-123" &&
							req.PaymentMethodID == "550e8400-e29b-41d4-a716-446655440000" &&
							len(req.PartnerConfig.VirtualAccount.Items) == 1
					})).Return(nil)
			},
		},
		{
			name:            "SUCCESS: Update VA range with both close and open ranges",
			expectedStatus:  http.StatusOK,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}, "openRange": {"binPrefix": "54321", "start": "100", "end": "500"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything,
					mock.MatchedBy(func(req *paymentMethodModel.SetupPaymentMethodConfigRequest) bool {
						return req.MerchantID == "merchant-123" &&
							req.PaymentMethodID == "550e8400-e29b-41d4-a716-446655440000" &&
							len(req.PartnerConfig.VirtualAccount.Items) == 2
					})).Return(nil)
			},
		},
		{
			name:            "SUCCESS: Update VA range with parent merchant",
			expectedStatus:  http.StatusOK,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID:      "merchant-123",
					MID:       sql.NullString{String: "MID123", Valid: true},
					ParentID:  sql.NullString{String: "parent-merchant-123", Valid: true},
					KYCStatus: sql.NullString{String: constant.KYCStatusNotRequired, Valid: true},
				}
				parentMerchant := &merchantModel.Merchant{
					UUID: "parent-merchant-123",
					MID:  sql.NullString{String: "PARENT_MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(parentMerchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything,
					mock.MatchedBy(func(req *paymentMethodModel.SetupPaymentMethodConfigRequest) bool {
						return req.MerchantID == "parent-merchant-123" &&
							req.PaymentMethodID == "550e8400-e29b-41d4-a716-446655440000" &&
							req.PartnerConfig.VirtualAccount.MerchantMID == "PARENT_MID123"
					})).Return(nil)
			},
		},
		{
			name:            "ERROR: Payment method service setup config error",
			expectedStatus:  http.StatusInternalServerError,
			paymentMethodID: "550e8400-e29b-41d4-a716-446655440000",
			requestBody:     `{"closeRange": {"binPrefix": "12345", "start": "001", "end": "999"}}`,
			hasUserContext:  true,
			setupMocks: func(merchantSvc *serviceMocks.IMerchantService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				merchant := &merchantModel.Merchant{
					UUID: "merchant-123",
					MID:  sql.NullString{String: "MID123", Valid: true},
				}
				merchantSvc.On("FindMerchantByID", mock.Anything, "merchant-123").Return(merchant, nil)

				paymentMethodSvc.On("SetupConfig", mock.Anything, mock.Anything).Return(assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantService := &serviceMocks.IMerchantService{}
			paymentMethodService := &serviceMocks.IPaymentMethodService{}
			tc.setupMocks(merchantService, paymentMethodService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithMerchantService(merchantService),
				WithPaymentMethodService(paymentMethodService),
			)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/payments/static-va/range/"+tc.paymentMethodID, strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)

				if tc.paymentMethodID != "" {
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("id", tc.paymentMethodID)
					ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
				}
				req = req.WithContext(ctx)
			}

			controller.UpdateVARange(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			merchantService.AssertExpectations(t)
			paymentMethodService.AssertExpectations(t)
		})
	}
}
