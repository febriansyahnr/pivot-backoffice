package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

func TestGetPaymentInsight(t *testing.T) {
	var (
		validUserClaims = &user.UserTokenClaims{
			UUID:       uuid.NewString(),
			MerchantId: uuid.NewString(),
		}
		mockPaymentService mockService.IPaymentService
		insightController  = PaymentController{
			config: &config.Config{
				ServiceName: "unit-test",
			},
			paymentService: &mockPaymentService,
		}
	)

	testCases := []struct {
		name           string
		callMock       func()
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:      "when everything is ok, should return 200",
			userClaim: validUserClaims,
			callMock: func() {
				mockPaymentService.On("GetTotalPaymentBalance", mock.Anything, constant.UuidMockType()).
					Return(&commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    "1945000.00",
					}, nil).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 3,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "1945000.00",
						},
					}, nil).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 0,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "0.00",
						},
					}, nil).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_VOID,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 1,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "10000.00",
						},
					}, nil).
					Once()

			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "when error occurred on get payment balance, then should return 500",
			userClaim: validUserClaims,
			callMock: func() {
				mockPaymentService.On("GetTotalPaymentBalance", mock.Anything, constant.UuidMockType()).
					Return(&commonModel.Amount{}, errors.New("merchant account not found")).Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 3,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "1945000.00",
						},
					}, nil).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 0,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "0.00",
						},
					}, nil).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_VOID,
				}).
					Return(&paymentModel.PaymentInsightItem{
						Total: 1,
						TotalAmount: commonModel.Amount{
							Currency: constant.CurrencyIDR,
							Value:    "10000.00",
						},
					}, nil).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "when error occurred on get total success insight, then should return 500",
			userClaim: validUserClaims,
			callMock: func() {
				mockPaymentService.On("GetTotalPaymentBalance", mock.Anything, constant.UuidMockType()).
					Return(nil, nil).Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("database error"))).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("database error"))).
					Once()

				mockPaymentService.On("GetTodayPaymentInsight", mock.Anything, paymentModel.PaymentInsightOption{
					MerchantID: validUserClaims.MerchantId,
					Status:     paymentConstant.PAYMENT_STATUS_VOID,
				}).
					Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("database error"))).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "when request does't have auth header, then should return 401",
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "when merchantID was invalid, then should return 401",
			userClaim: &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: "invalid-merchant-id",
			},
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/insights", nil)
			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(insightController.GetPaymentInsight)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestGetPaymentDashboardInsights(t *testing.T) {
	paymentService := mockService.NewIPaymentService(t)

	handler := New(nil, nil, nil, WithPaymentService(paymentService))

	router := chi.NewRouter()
	router.Get("/payment-insights", handler.GetPaymentDashboardInsights)

	now := time.Now().UTC()
	userClaims := &user.UserTokenClaims{}
	queryParams := fmt.Sprintf("insightStartDate=%s&insightEndDate=%s", now.AddDate(0, 0, -1).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))

	tests := []struct {
		name             string
		claims           *user.UserTokenClaims
		params           string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:User not found", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:             "ERROR:Invalid date range format", // NOSONAR
			claims:           userClaims,
			params:           "insightStartDate=ABC&insightEndDate=DEF", // NOSONAR
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "Key: insightStartDate Value: ABC Error: Value format must be yyyy-mm-ddThh:nn:ssZ"),
		},
		{
			name:             "ERROR:Empty date range", // NOSONAR
			claims:           userClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "start date or end date cannot be empty"),
		},
		{
			name:   "ERROR:Some error", // NOSONAR
			claims: userClaims,
			params: queryParams,
			setupMock: func() {
				paymentService.On("GetPaymentDashboardInsights", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: constant.WrapErrApiRespForTest(99, response.ErrTypeAPI, "an error occurred on the server. please try again later"),
		},
		{
			name:   "SUCCESS", // NOSONAR
			claims: userClaims,
			params: queryParams,
			setupMock: func() {
				paymentService.On("GetPaymentDashboardInsights", mock.Anything, mock.Anything).Once().Return(&paymentModel.PaymentDashboardInsights{
					WaitingForCaptureCount: 1, // NOSONAR
					SuccessRate:            &paymentModel.PaymentSuccessRateComparison{},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"waitingForCaptureCount":1,"paidCount":0,"paidTotal":0,"refundedCount":0,"refundedTotal":0,"failedCount":0,"failedTotal":0,"failedRefundCount":0,"successRate":{"previousSuccessRate":0,"currentSuccessRate":0,"differenceRate":0},"failureReasons":null}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payment-insights?"+test.params, nil)

			if test.claims != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), constant.CtxUserInfoKey, test.claims),
				)
			}
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Result:", rec.Body.String())
			}
		})
	}
}
