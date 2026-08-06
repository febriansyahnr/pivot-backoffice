package refund

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReceipt(t *testing.T) {
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
			name:      "SUCCESS: returns receipt URL",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetReceipt", mock.Anything, &refundModel.GetRefundReceiptRequest{
					RefundID:   validRefundID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(&refundModel.GetRefundReceiptResponse{
					ReceiptURL: "https://storage.googleapis.com/signed-url",
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"receiptUrl":"https://storage.googleapis.com/signed-url"}}`,
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
			name:           "ERROR: invalid UUID returns 400",
			userClaim:      validUserClaims,
			refundID:       "not-a-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"uuid is required and must be a valid UUID","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: refund not found returns 404",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetReceipt", mock.Anything, &refundModel.GetRefundReceiptRequest{
					RefundID:   validRefundID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgError.New(response.HttpErrNotFound, constant.ErrRefundNotFound)).Once()
			},
			expectedStatus: http.StatusNotFound,
			wantRespBody:   `{"code":"44","message":"refund not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: refund not success returns 400",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetReceipt", mock.Anything, &refundModel.GetRefundReceiptRequest{
					RefundID:   validRefundID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgError.New(response.HttpErrRequest, errors.New("receipt is only available for successful refunds"))).Once()
			},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"receipt is only available for successful refunds","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: service error returns 500",
			userClaim: validUserClaims,
			refundID:  validRefundID,
			setupMock: func() {
				mockRefundSvc.On("GetReceipt", mock.Anything, &refundModel.GetRefundReceiptRequest{
					RefundID:   validRefundID,
					MerchantID: validUserClaims.MerchantId,
				}).Return(nil, pkgError.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"failed to generate receipt","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/refunds/"+tc.refundID+"/receipt", nil)

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("uuid", tc.refundID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.GetReceipt)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
				t.Log("Expected:", tc.wantRespBody)
				t.Log("Actual:", rr.Body.String())
			}
		})
	}
}
