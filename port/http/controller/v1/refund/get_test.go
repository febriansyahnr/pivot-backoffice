package refund

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByID(t *testing.T) {
	var (
		validUserClaims = &userModel.UserTokenClaims{
			UUID:       "valid-user-id",
			MerchantId: "valid-merchant-id",
		}
		validRefundID = "019bbb37-4c6d-7bbb-bf5c-15478aa0b01a"
		mockRefundSvc = mockService.NewIRefundService(t)
		controller    = RefundController{
			refundService: mockRefundSvc,
		}
		fixedTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	)

	testCases := []struct {
		name           string
		userClaim      *userModel.UserTokenClaims
		refundID       string
		setupMock      func()
		expectedStatus int
		wantRespBody   string
	}{
		{
			name:      "SUCCESS: returns refund detail with status history",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetRefundDetailWithStatusHistories", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: validUserClaims.MerchantId,
					UUID:       validRefundID,
				}).Return(&refundModel.RefundResponse{
					ID:                validRefundID,
					ClientReferenceID: "client-ref-123",
					PaymentSessionID:  "payment-session-123",
					ChargeID:          "charge-123",
					Amount: commonModel.Amount{
						Value:    "50000",
						Currency: "IDR",
					},
					CapturedAmount: commonModel.Amount{
						Value:    "100000",
						Currency: "IDR",
					},
					IsFullAmount:    false,
					Status:          constant.RefundStatusSuccess,
					Reason:          constant.RefundReasonRequestedByCustomer,
					Description:     "Customer requested refund",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
					CreatedAt:       fixedTime,
					UpdatedAt:       fixedTime,
					StatusHistory: []unifiedPaymentModel.RefundStatusHistoryResponse{
						{
							Status:      constant.RefundStatusPending,
							Label:       "Refund Created",
							Description: "Refund has been created",
							Timestamp:   &fixedTime,
						},
						{
							Status:      constant.RefundStatusSuccess,
							Label:       "Success",
							Description: "Refund completed successfully",
							Timestamp:   &fixedTime,
						},
					},
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"019bbb37-4c6d-7bbb-bf5c-15478aa0b01a","clientReferenceId":"client-ref-123","paymentSessionId":"payment-session-123","chargeId":"charge-123","capturedAmount":{"value":"100000","currency":"IDR"},"isFullAmount":false,"amount":{"value":"50000","currency":"IDR"},"status":"SUCCESS","reason":"REQUESTED_BY_CUSTOMER","description":"Customer requested refund","destinationType":"CHANNEL","method":"AUTO","createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-01-15T10:30:00Z","statusHistory":[{"status":"PENDING","label":"Refund Created","description":"Refund has been created","timestamp":"2024-01-15T10:30:00Z"},{"status":"SUCCESS","label":"Success","description":"Refund completed successfully","timestamp":"2024-01-15T10:30:00Z"}]}}`,
		},
		{
			name:           "ERROR: user info missing from context returns 401",
			userClaim:      nil,
			refundID:       validRefundID,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: empty refund ID returns 400",
			userClaim:      validUserClaims,
			refundID:       "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"id is required","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: refund not found returns 404",
			userClaim: validUserClaims,
			refundID:  "00000000-0000-4000-8000-000000000000",
			setupMock: func() {
				mockRefundSvc.On("GetRefundDetailWithStatusHistories", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: validUserClaims.MerchantId,
					UUID:       "00000000-0000-4000-8000-000000000000",
				}).Return(nil, pkgError.New(response.HttpErrNotFound, constant.ErrRefundNotFound)).Once()
			},
			expectedStatus: http.StatusNotFound,
			wantRespBody:   `{"code":"44","message":"refund not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: service error returns 500",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetRefundDetailWithStatusHistories", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: validUserClaims.MerchantId,
					UUID:       validRefundID,
				}).Return(nil, pkgError.New(response.HttpErrDatabase, errors.New("database error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"98","message":"database error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "SUCCESS: returns refund detail without status history",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetRefundDetailWithStatusHistories", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: validUserClaims.MerchantId,
					UUID:       validRefundID,
				}).Return(&refundModel.RefundResponse{
					ID:                validRefundID,
					ClientReferenceID: "client-ref-456",
					PaymentSessionID:  "payment-session-456",
					ChargeID:          "charge-456",
					Amount: commonModel.Amount{
						Value:    "100000",
						Currency: "IDR",
					},
					CapturedAmount: commonModel.Amount{
						Value:    "100000",
						Currency: "IDR",
					},
					IsFullAmount:    true,
					Status:          constant.RefundStatusPending,
					Reason:          constant.RefundReasonDuplicate,
					Description:     "Duplicate payment",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
					CreatedAt:       fixedTime,
					UpdatedAt:       fixedTime,
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"019bbb37-4c6d-7bbb-bf5c-15478aa0b01a","clientReferenceId":"client-ref-456","paymentSessionId":"payment-session-456","chargeId":"charge-456","capturedAmount":{"value":"100000","currency":"IDR"},"isFullAmount":true,"amount":{"value":"100000","currency":"IDR"},"status":"PENDING","reason":"DUPLICATE","description":"Duplicate payment","destinationType":"CHANNEL","method":"AUTO","createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-01-15T10:30:00Z"}}`,
		},
		{
			name:      "SUCCESS: returns failed refund with failure code",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetRefundDetailWithStatusHistories", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: validUserClaims.MerchantId,
					UUID:       validRefundID,
				}).Return(&refundModel.RefundResponse{
					ID:                validRefundID,
					ClientReferenceID: "client-ref-789",
					PaymentSessionID:  "payment-session-789",
					ChargeID:          "charge-789",
					Amount: commonModel.Amount{
						Value:    "75000",
						Currency: "IDR",
					},
					CapturedAmount: commonModel.Amount{
						Value:    "100000",
						Currency: "IDR",
					},
					IsFullAmount:    false,
					Status:          constant.RefundStatusFailed,
					Reason:          constant.RefundReasonOthers,
					Description:     "Other reason",
					DestinationType: constant.RefundDestinationTypeChannel,
					Method:          constant.RefundMethodAuto,
					FailureCode:     "INSUFFICIENT_BALANCE",
					CreatedAt:       fixedTime,
					UpdatedAt:       fixedTime,
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"019bbb37-4c6d-7bbb-bf5c-15478aa0b01a","clientReferenceId":"client-ref-789","paymentSessionId":"payment-session-789","chargeId":"charge-789","capturedAmount":{"value":"100000","currency":"IDR"},"isFullAmount":false,"amount":{"value":"75000","currency":"IDR"},"status":"FAILED","reason":"OTHERS","description":"Other reason","destinationType":"CHANNEL","method":"AUTO","failureCode":"INSUFFICIENT_BALANCE","createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-01-15T10:30:00Z"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/refunds/"+tc.refundID, nil)

			if tc.refundID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("uuid", tc.refundID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			}

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			rr := httptest.NewRecorder()

			// Create handler and serve request
			handler := http.HandlerFunc(controller.GetByID)
			handler.ServeHTTP(rr, req)

			// Log response for debugging
			t.Logf("Response body: %s", rr.Body.String())

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
				t.Log("Expected:", tc.wantRespBody)
				t.Log("Actual:", rr.Body.String())
			}
		})
	}
}

func TestNewRefundController(t *testing.T) {
	mockRefundSvc := mockService.NewIRefundService(t)

	c := New(
		WithRefundService(mockRefundSvc),
	)

	assert.NotNil(t, c)
}
