package transfer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/transfer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	payloadRequest := &transfer.TransferRequest{
		RecipientID:  uuid.New().String(),
		ReferenceID:  "referenceId",
		Amount:       10,
		Remarks:      "test",
		TransferType: constant.MoneyFlowDirect,
	}

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
		requestBody  func() []byte
		setup        func(svc *mockSvc.ITransferService)
		setupHeader  func(req *http.Request)
		setupClaims  bool
	}{
		{
			name: "SUCCESS: Merchant Transfer",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.ITransferService) {
				svc.On("Transfer", mock.Anything, mock.Anything).Return(trsfr, nil)
			},
			setupClaims:  true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceId":"referenceId","transferType":"INDIRECT","amount":10,"status":"PENDING","remarks":"test remarks","updatedAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "SUCCESS: Sub Merchant Transfer",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.ITransferService) {
				svc.On("Transfer", mock.Anything, mock.Anything).Return(trsfr, nil)
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceId":"referenceId","transferType":"INDIRECT","amount":10,"status":"PENDING","remarks":"test remarks","updatedAt":"0001-01-01T00:00:00Z","createdAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "ERROR: Unable get claims",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup:        func(_ *mockSvc.ITransferService) { /* No Body */ },
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","errors":"invalid token"}`,
			setupClaims:  false,
		},
		{
			name: "ERROR: Failed decode request",
			requestBody: func() []byte {
				return []byte("invalid")
			},
			setup:        func(svc *mockSvc.ITransferService) { /* No Body */ },
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name:       "ERROR: Failed decode request (new error response)",
			merchantId: "93a58f2b-72b6-4bef-9906-f691c6e68158", // NOSONAR
			requestBody: func() []byte {
				return []byte("invalid")
			},
			setup:        func(svc *mockSvc.ITransferService) { /* No Body */ },
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"api_validation_error","message":"The request was invalid, or an error occurred in downstream provider","error":{"type":"API_ERROR","details":[{"field":"","message":"The request was invalid, or an error occurred in downstream provider"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Payload failed validation",
			requestBody: func() []byte {
				modifiedRequest := *payloadRequest
				modifiedRequest.Amount = 0
				modifiedRequest.RecipientID = ""
				payloadRequestByte, _ := json.Marshal(modifiedRequest)
				return payloadRequestByte
			},
			setup:        func(svc *mockSvc.ITransferService) { /* No Body Function */ },
			setupHeader:  func(req *http.Request) { req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String()) },
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","errors":{"Amount":"Key: 'TransferRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag","RecipientID":"Key: 'TransferRequest.RecipientID' Error:Field validation for 'RecipientID' failed on the 'required' tag"}}`,
		},
		{
			name:       "ERROR: Payload failed validation (new error response)",
			merchantId: "93a58f2b-72b6-4bef-9906-f691c6e68158", // NOSONAR
			requestBody: func() []byte {
				req := *payloadRequest
				req.ReferenceID = strings.Repeat("TEST", 100)
				raw, _ := json.Marshal(req)
				return raw
			},
			setup:        func(svc *mockSvc.ITransferService) { /* No Body Function */ },
			setupHeader:  func(req *http.Request) { req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String()) },
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"api_validation_error","message":"The request was invalid, or an error occurred in downstream provider","error":{"type":"API_ERROR","details":[{"field":"referenceID","message":"Make sure the number of characters in the referenceID column is not greater than 100"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Transfer",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.ITransferService) {
				svc.On("Transfer", mock.Anything, mock.Anything).Return(nil, constant.ErrSameMerchant)
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","errors":"cannot transfer to the same merchant"}`,
		},
		{
			name:       "ERROR: Transfer (new error response)",
			merchantId: "93a58f2b-72b6-4bef-9906-f691c6e68158", // NOSONAR
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.ITransferService) {
				svc.On("Transfer", mock.Anything, mock.Anything).Return(nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrSameMerchant))
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: `{"code":"unprocessable_entity","message":"cannot transfer to the same merchant","error":{"type":"API_ERROR","details":[{"field":"","message":"cannot transfer to the same merchant"}],"traceId":""}}`,
		},
		{
			name:       "ERROR: Insufficient balance (new error response)",
			merchantId: "93a58f2b-72b6-4bef-9906-f691c6e68158", // NOSONAR
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.ITransferService) {
				svc.On("Transfer", mock.Anything, mock.Anything).Return(nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrInsufficientBalance))
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			expectedCode: http.StatusForbidden,
			expectedBody: `{"code":"balance_insufficient","message":"Merchant Balance is Insufficient","error":{"type":"API_ERROR","details":[{"field":"","message":"Re Top-up your Balance"}],"traceId":""}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := mockSvc.NewITransferService(t)
			vld := validatorExt.New()
			tc.setup(svc)

			ctrl := New(svc, vld)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(tc.requestBody()))

			if tc.merchantId == "" {
				tc.merchantId = merchantPlatformWhitelistedOldResponseFormat
			}

			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, tc.merchantId)
			if tc.setupClaims {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{MerchantId: tc.merchantId})
			}
			req = req.WithContext(ctx)

			if tc.setupHeader != nil {
				tc.setupHeader(req)
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.Create)
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
