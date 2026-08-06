package payment_test

import (
	"context"
	"fmt"
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
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"
	"github.com/paper-indonesia/pdk/go/monitoring"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFilterStaticQrisList(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		queryParams    string
		hasUserContext bool
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			queryParams:    "?page=1&perPage=10",
			hasUserContext: false,
		},
		{
			name:           "ERROR: Invalid page parameter",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=invalid&perPage=10",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid perPage parameter",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=invalid",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid startDate format",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=10&startDate=invalid-date",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid endDate format",
			expectedStatus: http.StatusBadRequest,
			queryParams:    "?page=1&perPage=10&endDate=invalid-date",
			hasUserContext: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-qris"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
				req = req.WithContext(ctx)
			}

			controller.FilterStaticQrisList(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetStaticQrisDetail(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		hasUserContext bool
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			hasUserContext: false,
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			hasUserContext: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-qris/"+tc.paymentID, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("paymentId", tc.paymentID)
				ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
				req = req.WithContext(ctx)
			}

			controller.GetStaticQrisDetail(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetStaticQrisTransactions(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		queryParams    string
		hasUserContext bool
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=10",
			hasUserContext: false,
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			queryParams:    "?page=1&perPage=10",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid page parameter",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			queryParams:    "?page=invalid&perPage=10",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid perPage parameter",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=invalid",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid startDate format",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=10&startDate=invalid-date",
			hasUserContext: true,
		},
		{
			name:           "ERROR: Invalid endDate format",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			queryParams:    "?page=1&perPage=10&endDate=invalid-date",
			hasUserContext: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-qris/"+tc.paymentID+"/transactions"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("paymentId", tc.paymentID)
				ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
				req = req.WithContext(ctx)
			}

			controller.GetStaticQrisTransactions(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestDeactivateStaticQris(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name           string
		expectedStatus int
		paymentID      string
		requestBody    string
		hasUserContext bool
		hasXRequestPin bool
		userUUID       string
		setupMocks     func(*serviceMocks.IPaymentService, *serviceMocks.IUserService)
	}{
		{
			name:           "ERROR: Missing user context",
			expectedStatus: http.StatusUnauthorized,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: false,
			hasXRequestPin: true,
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				// No mocks needed for this test case
			},
		},
		{
			name:           "ERROR: Missing x-request-pin header",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasXRequestPin: false,
			userUUID:       "user-123",
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				// No mocks needed for this test case
			},
		},
		{
			name:           "ERROR: Empty payment ID",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasXRequestPin: true,
			userUUID:       "user-123",
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				userSvc.On("CheckCurrentPin", mock.Anything, "user-123", "test").Return(nil)
			},
		},
		{
			name:           "ERROR: Invalid JSON body",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `{"status": }`,
			hasUserContext: true,
			hasXRequestPin: true,
			userUUID:       "user-123",
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				userSvc.On("CheckCurrentPin", mock.Anything, "user-123", "test").Return(nil)
			},
		},
		{
			name:           "ERROR: Missing status field",
			expectedStatus: http.StatusBadRequest,
			paymentID:      "payment-123",
			requestBody:    `{}`,
			hasUserContext: true,
			hasXRequestPin: true,
			userUUID:       "user-123",
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				userSvc.On("CheckCurrentPin", mock.Anything, "user-123", "test").Return(nil)
			},
		},
		{
			name:           "SUCCESS: Update static QRIS status",
			expectedStatus: http.StatusOK,
			paymentID:      "payment-123",
			requestBody:    `{"status": "INACTIVE"}`,
			hasUserContext: true,
			hasXRequestPin: true,
			userUUID:       "user-123",
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService, userSvc *serviceMocks.IUserService) {
				userSvc.On("CheckCurrentPin", mock.Anything, "user-123", "test").Return(nil)
				paymentSvc.On("DeactivateStaticQris", mock.Anything, "payment-123", "merchant-123",
					mock.MatchedBy(func(req paymentModel.StaticQrisUpdateStatusRequest) bool {
						return req.Status == "INACTIVE"
					})).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := &serviceMocks.IPaymentService{}
			userService := &serviceMocks.IUserService{}
			tc.setupMocks(paymentService, userService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
				WithUserService(userService),
			)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/payments/static-qris/"+tc.paymentID+"/status", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tc.hasXRequestPin {
				req.Header.Set("x-request-pin", "dGVzdA==")
			}

			w := httptest.NewRecorder()

			if tc.hasUserContext {
				userClaims := &userModel.UserTokenClaims{
					UUID:       tc.userUUID,
					MerchantId: "merchant-123",
				}
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)

				if tc.paymentID != "" {
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("paymentId", tc.paymentID)
					ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
				}
				req = req.WithContext(ctx)
			}

			controller.DeactivateStaticQris(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			paymentService.AssertExpectations(t)
			userService.AssertExpectations(t)
		})
	}
}

func TestGetMaxActiveStaticQRPerMerchant(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name                     string
		expectedStatus           int
		expectedMaxActiveQRLimit int
		setupMocks               func(*serviceMocks.IPaymentService)
	}{
		{
			name:                     "SUCCESS: Get max active static QR per merchant - returns 12",
			expectedStatus:           http.StatusOK,
			expectedMaxActiveQRLimit: 12,
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService) {
				paymentSvc.On("GetMaxActiveStaticQRPerMerchant").Return(12)
			},
		},
		{
			name:                     "SUCCESS: Get max active static QR per merchant - returns 0",
			expectedStatus:           http.StatusOK,
			expectedMaxActiveQRLimit: 0,
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService) {
				paymentSvc.On("GetMaxActiveStaticQRPerMerchant").Return(0)
			},
		},
		{
			name:                     "SUCCESS: Get max active static QR per merchant - returns 100",
			expectedStatus:           http.StatusOK,
			expectedMaxActiveQRLimit: 100,
			setupMocks: func(paymentSvc *serviceMocks.IPaymentService) {
				paymentSvc.On("GetMaxActiveStaticQRPerMerchant").Return(100)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentService := &serviceMocks.IPaymentService{}
			tc.setupMocks(paymentService)

			controller := New(
				&config.Config{},
				validator.New(),
				&monitoring.Monitor{},
				WithLogger(logger),
				WithPaymentService(paymentService),
			)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/static-qris/max-active-limit", nil)
			w := httptest.NewRecorder()

			controller.GetMaxActiveStaticQRPerMerchant(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				expectedBody := `{"code":"00","data":{"maxActiveStaticQRPerMerchant":` + 
					fmt.Sprintf("%d", tc.expectedMaxActiveQRLimit) + 
					`},"message":"OK"}`
				assert.JSONEq(t, expectedBody, w.Body.String())
			}

			paymentService.AssertExpectations(t)
		})
	}
}
