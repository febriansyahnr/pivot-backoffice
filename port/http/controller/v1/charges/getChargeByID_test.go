package charges

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChargeByID(t *testing.T) {
	var (
		validUserClaims = &userModel.UserTokenClaims{
			UUID:       "valid-user-id",
			MerchantId: "valid-merchant-id",
		}
		validChargeID         = "4b9c6c20-e885-4122-83a3-c165042f4aaa"
		mockUnifiedPaymentSvc = mockService.NewIUnifiedPaymentService(t)
		mockMerchantService   = mockService.NewIMerchantService(t)
		controller            = ChargesController{
			unifiedPaymentService: mockUnifiedPaymentSvc,
			merchantService:       mockMerchantService,
		}
	)

	testCases := []struct {
		name           string
		callMock       func()
		expectedStatus int
		wantRespBody   string
		userClaim      *userModel.UserTokenClaims
		chargeID       string
		queryParams    string
	}{
		{
			name:      "when everything is ok, should return 200",
			userClaim: validUserClaims,
			chargeID:  validChargeID,
			callMock: func() {
				mockUnifiedPaymentSvc.On("GetChargeDetail", mock.Anything, &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
					ChargeID:   validChargeID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(&unifiedPaymentModel.ChargeResponse{
					ID:               validChargeID,
					PaymentSessionID: "payment-session-id",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					Status:    constant.ChargeStatusSuccess,
					CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
					PaidAt:    &[]time.Time{time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)}[0],
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"4b9c6c20-e885-4122-83a3-c165042f4aaa","paymentSessionId":"payment-session-id","paymentSessionClientReferenceId":"","amount":{"value":100000,"currency":"IDR"},"statementDescriptor":"","status":"SUCCESS","authorizedAmount":null,"capturedAmount":null,"isCaptured":false,"createdAt":"2024-01-01T10:00:00Z","updatedAt":"2024-01-01T10:30:00Z","paidAt":"2024-01-01T10:30:00Z"}}`,
		},
		{
			name:      "when charge detail retrieval fails, should return 500",
			userClaim: validUserClaims,
			chargeID:  validChargeID,
			callMock: func() {
				mockUnifiedPaymentSvc.On("GetChargeDetail", mock.Anything, &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
					ChargeID:   validChargeID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgError.New("internal server error", errors.New("internal service error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"internal service error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when user info is missing from context, should return 401",
			userClaim:      nil,
			chargeID:       validChargeID,
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when chargeID is missing, should return 400",
			userClaim:      validUserClaims,
			chargeID:       "",
			callMock:       func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when chargeID is invalid UUID format, should return 400",
			userClaim:      validUserClaims,
			chargeID:       "invalid-uuid-format",
			callMock:       func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "when the subMerchantId is invalid, should return 400",
			userClaim:   validUserClaims,
			chargeID:    validChargeID,
			queryParams: "subMerchantId=invalid-sub-merchant-id",
			callMock: func() {
				mockMerchantService.On("ValidateSubMerchantParent", mock.Anything, validUserClaims.MerchantId, "invalid-sub-merchant-id").Return(pkgError.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchantParent)).Once()
			},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"incorrect submerchant parent","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "when the subMerchantId is valid, should return 200",
			userClaim:   validUserClaims,
			chargeID:    validChargeID,
			queryParams: "subMerchantId=valid-sub-merchant-id",
			callMock: func() {
				mockMerchantService.On("ValidateSubMerchantParent", mock.Anything, validUserClaims.MerchantId, "valid-sub-merchant-id").Return(nil).Once()
				mockUnifiedPaymentSvc.On("GetChargeDetail", mock.Anything, &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
					ChargeID:   validChargeID,
					MerchantID: "valid-sub-merchant-id", // Should use sub-merchant ID
				}).Return(&unifiedPaymentModel.ChargeResponse{
					ID:               validChargeID,
					PaymentSessionID: "payment-session-id",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					Status:    constant.ChargeStatusSuccess,
					CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
					PaidAt:    &[]time.Time{time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)}[0],
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"4b9c6c20-e885-4122-83a3-c165042f4aaa","paymentSessionId":"payment-session-id","paymentSessionClientReferenceId":"","amount":{"value":100000,"currency":"IDR"},"statementDescriptor":"","status":"SUCCESS","authorizedAmount":null,"capturedAmount":null,"isCaptured":false,"createdAt":"2024-01-01T10:00:00Z","updatedAt":"2024-01-01T10:30:00Z","paidAt":"2024-01-01T10:30:00Z"}}`,
		},
		{
			name:      "when charge is not found, should return 404",
			userClaim: validUserClaims,
			chargeID:  "00000000-0000-4000-8000-000000000000", // Valid UUID format but non-existent
			callMock: func() {
				mockUnifiedPaymentSvc.On("GetChargeDetail", mock.Anything, &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
					ChargeID:   "00000000-0000-4000-8000-000000000000",
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgError.New(response.HttpErrNotFound, constant.ErrDataNotFound)).Once()
			},
			expectedStatus: http.StatusNotFound,
			wantRespBody:   `{"code":"44","message":"data not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "when charge status is failed with failure details, should return charge with failure info",
			userClaim: validUserClaims,
			chargeID:  validChargeID,
			callMock: func() {
				mockUnifiedPaymentSvc.On("GetChargeDetail", mock.Anything, &unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
					ChargeID:   validChargeID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(&unifiedPaymentModel.ChargeResponse{
					ID:               validChargeID,
					PaymentSessionID: "payment-session-id",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    100000.00,
					},
					Status:         constant.ChargeStatusFailed,
					FailureCode:    "DECLINED_BY_CHANNEL",
					FailureMessage: "The transaction was declined by the channel.",
					Recommendation: "The cardholder should contact their issuer for clarification.",
					CreatedAt:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt:      time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"4b9c6c20-e885-4122-83a3-c165042f4aaa","paymentSessionId":"payment-session-id","paymentSessionClientReferenceId":"","amount":{"value":100000,"currency":"IDR"},"statementDescriptor":"","status":"FAILED","authorizedAmount":null,"capturedAmount":null,"isCaptured":false,"failureCode":"DECLINED_BY_CHANNEL","failureMessage":"The transaction was declined by the channel.","recommendation":"The cardholder should contact their issuer for clarification.","createdAt":"2024-01-01T10:00:00Z","updatedAt":"2024-01-01T10:30:00Z","paidAt":null}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/charges/%s?%s", tc.chargeID, tc.queryParams), nil)

			if tc.chargeID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("uuid", tc.chargeID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			}

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.GetChargeByID)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
				t.Log("Expected:", tc.wantRespBody)
				t.Log("Actual:", rr.Body.String())
			}
		})
	}
}

func TestNewChargesController(t *testing.T) {
	var (
		mockUnifiedPaymentService = new(mockService.IUnifiedPaymentService)
		mockMerchantService       = new(mockService.IMerchantService)
		mockLogger                logger.ILogger
	)

	c := New(&config.Config{}, validator.New(), &monitoring.Monitor{},
		WithLogger(mockLogger),
		WithUnifiedPaymentService(mockUnifiedPaymentService),
		WithMerchantService(mockMerchantService),
	)

	assert.NotNil(t, c)
	assert.NotNil(t, c.(*ChargesController).unifiedPaymentService)
	assert.NotNil(t, c.(*ChargesController).merchantService)
}
