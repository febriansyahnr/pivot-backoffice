package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionList(t *testing.T) {
	userClaims := &userModel.UserTokenClaims{
		Name:       "user",
		MerchantId: uuid.Nil.String(),
	}
	testCases := []struct {
		name         string
		setup        func(platformSvc *mocks.IPlatformService)
		setupClaims  func(r *http.Request) *http.Request
		setupParams  func() string
		expectedCode int
		expectedBody string
	}{
		{
			name: "SUCCESS: Get Merchant Transfer Transaction List",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
				platformSvc.On(
					"GetMerchantTransactionList",
					mock.Anything,
					mock.Anything,
				).Return(
					&commonModel.PaginationResponse{
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 100,
							TotalPages: 10,
						},
						Data: []transfer.ListTransferResponse{
							{
								UUID:          uuid.Nil.String(),
								ReferenceID:   "referenceId",
								SenderID:      uuid.Nil.String(),
								SenderName:    "senderName",
								RecipientID:   uuid.Max.String(),
								RecipientName: "recipientName",
								TransferType:  "transferType",
								Amount:        1000,
								Status:        "SUCCESS",
								Remarks:       "remarks",
							},
						},
					},
					nil,
				)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"OK","data":{"data":[{"uuid":"00000000-0000-0000-0000-000000000000","referenceId":"referenceId","type":"","senderId":"00000000-0000-0000-0000-000000000000","senderName":"senderName","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientName":"recipientName","transferType":"transferType","amount":1000,"status":"SUCCESS", "remarks":"remarks","updatedAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z"}],"meta":{"page":1,"perPage":10,"totalItems":100,"totalPages":10}}}`,
		},
		{
			name: "ERROR: Missing claims",
			setupClaims: func(r *http.Request) *http.Request {
				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"invalid access","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid StartDate",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "x")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "x")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid startDate value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Enddate",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "x")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "x")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid endDate value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid PaymentStartDate",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("paymentStartDate", "x")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid paymentStartDate value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid PaymentEndDate",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("paymentStartDate", "")
				params.Add("paymentEndDate", "x")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid paymentEndDate value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid date range values",
			setupClaims: func(r *http.Request) *http.Request {
				return r.WithContext(
					context.WithValue(r.Context(), constant.CtxUserInfoKey, userClaims),
				)
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("startDate", "2025-01-01T17:00:00.000Z")
				params.Add("endDate", "2025-01-31T16:59:59.999Z")

				return params.Encode()
			},
			setup:        func(*mocks.IPlatformService) { /* No Body */ },
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months.","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Page",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "x")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid page value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Page Size",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "")
				params.Add("perPage", "x")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid perPage value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Error Get Merchant Transfer Transaction List",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("reference", constant.ReferencePlatformTransfer)
				params.Add("referenceType", "type")
				params.Add("referenceId", "refId")
				params.Add("status", "")
				params.Add("approvalStatus", "")
				params.Add("paymentMethod", "")
				params.Add("keyword", "")
				params.Add("startDate", "")
				params.Add("endDate", "")
				params.Add("sortBy", "")
				params.Add("sort", "")
				params.Add("page", "")
				params.Add("perPage", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
				platformSvc.On(
					"GetMerchantTransactionList",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			svc := mocks.NewIPlatformService(t)
			tc.setup(svc)

			ctrl := New(logger, nil, svc)

			urlParams := tc.setupParams()
			req := httptest.NewRequest(http.MethodGet, "/transactions?"+urlParams, nil)
			req = tc.setupClaims(req)

			rr := httptest.NewRecorder()
			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.TransactionList)
			handler.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())

		})
	}
}
