package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentHistory(t *testing.T) {
	var (
		validUserClaims = &userModel.UserTokenClaims{
			UUID:       "valid-user-id",
			MerchantId: "valid-merchant-id",
		}
		paymentID           = "valid-payment-id"
		mockPaymentService  = mockService.NewIPaymentService(t)
		mockMerchantService = mockService.NewIMerchantService(t)
		controller          = PaymentController{
			paymentService:  mockPaymentService,
			merchantService: mockMerchantService,
		}
	)

	testCases := []struct {
		name           string
		callMock       func()
		expectedStatus int
		wantRespBody   string
		userClaim      *userModel.UserTokenClaims
		paymentID      string
		queryParams    string
	}{
		{
			name:      "when everything is ok, should return 200",
			userClaim: validUserClaims,
			paymentID: paymentID,
			callMock: func() {
				mockPaymentService.On("GetPaymentHistoryDetail", mock.Anything, paymentModel.PaymentHistoryDetailOption{
					PaymentID:  paymentID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(&paymentModel.PaymentHistoryDetailResponse{
					UUID:            paymentID,
					MerchantID:      validUserClaims.MerchantId,
					Amount:          commonModel.Amount{Currency: "IDR", Value: "100000.00"},
					Status:          "success",
					SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"valid-payment-id","merchantId":"valid-merchant-id","customerId":"","referenceId":"","paymentMethod":"","paymentMethodType":"","processorReferenceNumber":"","paymentTypeDetail":{},"paymentChannel":"","bankReferenceId":"","amount":{"currency":"IDR","value":"100000.00"},"amountPaid":{"currency":"","value":""},"status":"success","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","expiredAt":null,"investigationStartedAt":null,"cancelledAt":null,"totalSplitAmount":{"currency":"","value":""},"fee":{"currency":"","value":""},"settledAmount":{"currency":"","value":""},"settlementModel":"AGGREGATOR","paymentType":""}}`,
		},
		{
			name:      "when payment history retrieval fails, should return 500",
			userClaim: validUserClaims,
			paymentID: paymentID,
			callMock: func() {
				mockPaymentService.On("GetPaymentHistoryDetail", mock.Anything, paymentModel.PaymentHistoryDetailOption{
					PaymentID:  paymentID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgErrors.New("internal server error", errors.New("internal service error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"internal service error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when user info is missing from context, should return 401",
			userClaim:      nil,
			paymentID:      paymentID,
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when paymentID is missing, should return 400",
			userClaim:      validUserClaims,
			paymentID:      "",
			callMock:       func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "when the subMerchantId is invalid, should return 400",
			userClaim:   validUserClaims,
			paymentID:   "valid-payment-id",
			queryParams: "subMerchantId=invalid-sub-merchant-id",
			callMock: func() {
				mockMerchantService.On("ValidateSubMerchantParent", mock.Anything, validUserClaims.MerchantId, "invalid-sub-merchant-id").Return(pkgErrors.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchantParent)).Once()
			},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"incorrect submerchant parent","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "when the subMerchantId is valid, should return 200",
			userClaim:   validUserClaims,
			paymentID:   paymentID,
			queryParams: "subMerchantId=valid-sub-merchant-id",
			callMock: func() {
				mockMerchantService.On("ValidateSubMerchantParent", mock.Anything, validUserClaims.MerchantId, "valid-sub-merchant-id").Return(nil).Once()
				mockPaymentService.On("GetPaymentHistoryDetail", mock.Anything, paymentModel.PaymentHistoryDetailOption{
					PaymentID:  paymentID,
					MerchantID: "valid-sub-merchant-id",
				}).Return(&paymentModel.PaymentHistoryDetailResponse{
					UUID:            paymentID,
					MerchantID:      validUserClaims.MerchantId,
					Amount:          commonModel.Amount{Currency: "IDR", Value: "100000.00"},
					Status:          "success",
					SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"valid-payment-id","merchantId":"valid-merchant-id","customerId":"","referenceId":"","paymentMethod":"","paymentMethodType":"","processorReferenceNumber":"","paymentTypeDetail":{},"paymentChannel":"","bankReferenceId":"","amount":{"currency":"IDR","value":"100000.00"},"amountPaid":{"currency":"","value":""},"status":"success","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","expiredAt":null,"investigationStartedAt":null,"cancelledAt":null,"totalSplitAmount":{"currency":"","value":""},"fee":{"currency":"","value":""},"settledAmount":{"currency":"","value":""},"settlementModel":"AGGREGATOR","paymentType":""}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/payments/%s?%s", tc.paymentID, tc.queryParams), nil)

			if tc.paymentID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("payment_id", tc.paymentID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			}

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.PaymentHistory)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
				t.Log("Result:", rr.Body.String())
			}
		})
	}
}

func TestNewPayment(t *testing.T) {
	var (
		mockPaymentService = new(mockService.IPaymentService)
		mockLogger         logger.ILogger
	)

	c := New(&config.Config{}, validator.New(), &monitoring.Monitor{}, WithLogger(mockLogger), WithPaymentService(mockPaymentService))
	assert.NotNil(t, c)
}

func TestGetEncryptionKey(t *testing.T) {
	paymentService := mockService.NewIPaymentService(t)

	handler := New(nil, nil, nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/encryption-key", handler.GetEncryptionKey)

	tests := []struct {
		name             string
		userInfo         *userModel.UserTokenClaims
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR: User not found", // NOSONAR
			setupMock:        func() { /* Empty Function */ },
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:     "ERROR: Some error", // NOSONAR
			userInfo: &userModel.UserTokenClaims{},
			setupMock: func() {
				paymentService.On("GetEncryptionKey", mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:     "SUCCESS", // NOSONAR
			userInfo: &userModel.UserTokenClaims{},
			setupMock: func() {
				paymentService.On("GetEncryptionKey", mock.Anything).Once().Return(&paymentModel.GetEncryptionKeyResponse{EncryptionKey: "ZW5jcnlwdGlvbi1rZXk="}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"encryptionKey":"ZW5jcnlwdGlvbi1rZXk="}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/encryption-key", nil)

			if test.userInfo != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userInfo))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}

			paymentService.AssertExpectations(t)
		})
	}
}
