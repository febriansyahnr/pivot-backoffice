package transfer_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/transfer"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetById(t *testing.T) {

	trsfr := &transfer.Transfer{
		UUID:         uuid.Max,
		MerchantID:   uuid.Max,
		RecipientID:  uuid.Max,
		ReferenceID:  "referenceId",
		TransferType: constant.MoneyFlowIndirect,
		Currency:     constant.CurrencyIDR,
		Amount:       10,
		Status:       constant.TransferStatusPending,
		Remarks:      "test remarks",
	}

	testCases := []struct {
		name         string
		merchantId   string
		expectedCode int
		expectedBody string
		setup        func(svc *mockSvc.ITransferService)
		setupHeader  func(req *http.Request)
		setupClaims  bool
		setupParam   bool
	}{
		{
			name: "SUCCESS: Get Transfer By ID",
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(trsfr, nil)
			},
			setupClaims:  true,
			setupParam:   true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceId":"referenceId","transferType":"INDIRECT","amount":10,"status":"PENDING","remarks":"test remarks","updatedAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "SUCCESS: Sub Merchant Get Transfer By ID",
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(trsfr, nil)
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			setupParam:   true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceId":"referenceId","transferType":"INDIRECT","amount":10,"status":"PENDING","remarks":"test remarks","updatedAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:         "ERROR: Unable get claims",
			setup:        func(_ *mockSvc.ITransferService) { /* No Body Function */ },
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"invalid token","error":{"type":"API_ERROR","message":"invalid token","recommendation":""},"data":null}`,
			setupParam:   true,
			setupClaims:  false,
		},
		{
			name:         "ERROR: Invalid ID",
			setup:        func(_ *mockSvc.ITransferService) { /* No Body Function */ },
			setupClaims:  true,
			setupParam:   false,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid id","error":{"type":"API_ERROR","message":"invalid id","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Get Transfer By ID",
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("errors"))
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			setupParam:   true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","message":"errors","recommendation":""},"data":null}`,
		},
		{
			name:       "ERROR: Transfer not found (new error response)",
			merchantId: "b1411d3d-a2ad-4409-bffc-178b34528929", // NOSONAR
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrTransferNotFound)
			},
			setupHeader:  func(req *http.Request) { req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String()) },
			setupClaims:  true,
			setupParam:   true,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":"resource_missing","message":"The transfer with ID 61de320f-6242-46db-8da4-f323ebc5f277 cannot be found","error":{"type":"GATEWAY_ERROR","details":[{"field":"","message":"The transfer with ID 61de320f-6242-46db-8da4-f323ebc5f277 cannot be found"}],"traceId":""}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := mockSvc.NewITransferService(t)
			tc.setup(svc)

			ctrl := New(svc, validatorExt.New())

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/transfers", nil)

			if tc.merchantId == "" {
				tc.merchantId = merchantPlatformWhitelistedOldResponseFormat
			}
			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, tc.merchantId)
			if tc.setupClaims {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
					MerchantId: merchantPlatformWhitelistedOldResponseFormat,
				})
			}
			req = req.WithContext(ctx)

			if tc.setupParam {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("id", "61de320f-6242-46db-8da4-f323ebc5f277")
				ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
				req = req.WithContext(ctx)
			}

			if tc.setupHeader != nil {
				tc.setupHeader(req)
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.GetById)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())

		})
	}
}
