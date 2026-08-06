package transfer_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/transfer"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {

	newMerchantId := "3bc07166-0b88-4794-8bb3-cbb6257b2023" // NOSONAR

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
	resp := &commonModel.PaginationResponse{
		Data: []*transfer.Transfer{trsfr},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name         string
		merchantId   string
		expectedCode int
		expectedBody string
		setup        func(svc *mockSvc.ITransferService)
		setupHeader  func(req *http.Request)
		setupParam   func(req *http.Request)
		setupClaims  bool
	}{
		{
			name: "SUCCESS: Get Transfer List",
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetList",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(resp, nil)
			},
			setupClaims: true,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=REF0001&page=1&perPage=10&perPage=10&startDate=2025-05-19T00:00:00%2B07:00&endDate=2025-05-19T23:59:59%2B07:00"
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":[{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","recipientId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceId":"referenceId","transferType":"INDIRECT","currency":"IDR","amount":10,"status":"PENDING","remarks":"test remarks","reasonDescription":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","deletedAt":null,"beneficiary":""}],"pagination":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}`,
		},
		{
			name: "ERROR: Unable get claims",
			setup: func(_ *mockSvc.ITransferService) {
				// No Body Function
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"invalid token","error":{"type":"API_ERROR","message":"invalid token","recommendation":""},"data":null}`,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=referenceId&page=1&perPage=10"
			},
			setupClaims: false,
		},
		{
			name: "ERROR: Invalid date range",
			setup: func(_ *mockSvc.ITransferService) {
				// No Body Function
			},
			setupClaims: true,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=referenceId&page=0&perPage=10&startDate=2025-05-19T00:00:00%2B07:00"
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"Key: 'GetTransferListRequest.EndDate' Error:Field validation for 'EndDate' failed on the 'required_with' tag","error":{"type":"API_ERROR","message":"Key: 'GetTransferListRequest.EndDate' Error:Field validation for 'EndDate' failed on the 'required_with' tag","recommendation":""},"data":null}`,
		},
		{
			name:       "ERROR: Invalid date range (new error response)",
			merchantId: newMerchantId,
			setup: func(_ *mockSvc.ITransferService) {
				// No Body Function
			},
			setupClaims: true,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=referenceId&page=0&perPage=10&startDate=2025-05-19T00:00:00%2B07:00"
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"api_validation_error","message":"The request was invalid, or an error occurred in downstream provider","error":{"type":"API_ERROR","details":[{"field":"endDate","message":"Make sure endDate value is fulfilled"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Invalid Page",
			setup: func(_ *mockSvc.ITransferService) {
				// No Body Function
			},
			setupClaims: true,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=referenceId&page=0&perPage=10"
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid page number","error":{"type":"API_ERROR","message":"invalid page number","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid PageSize",
			setup: func(_ *mockSvc.ITransferService) {
				// No Body Function
			},
			setupClaims: true,
			setupParam: func(req *http.Request) {
				req.URL.RawQuery = "referenceId=referenceId&page=1&perPage=0"
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid per page number","error":{"type":"API_ERROR","message":"invalid per page number","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Get Transfer List",
			setup: func(svc *mockSvc.ITransferService) {
				svc.On(
					"GetList",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil, errors.New("errors"))
			},
			setupHeader: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			setupClaims:  true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","message":"errors","recommendation":""},"data":null}`,
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

			if tc.setupHeader != nil {
				tc.setupHeader(req)
			}
			if tc.setupParam != nil {
				tc.setupParam(req)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.GetList)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			if !assert.JSONEq(t, tc.expectedBody, rr.Body.String()) {
				t.Log("Result:", rr.Body.String())
			}
		})
	}
}
