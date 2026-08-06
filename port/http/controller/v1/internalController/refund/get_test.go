package v1InternalRefundController_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/refund"
	"github.com/stretchr/testify/assert"
)

func TestGetList(t *testing.T) {
	refundSvc := serviceMocks.NewIRefundService(t)
	cfg := &config.Config{}
	controller := New(cfg, WithRefundService(refundSvc))

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR: Service returns error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundList", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS - Submerchant",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundList", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(&commonModel.PaginationResponse{}, nil)
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: uuid.NewString(),
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success", "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundList", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(&commonModel.PaginationResponse{}, nil)
			},
			requestHeader: nil,
			wantStatus:    http.StatusOK,
			wantResponse:  `{"code":"00", "data":null, "message":"Success", "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/refunds?page=1&perPage=5&startDate=2025-01-01T00:00:00Z&endDate=2025-01-01T00:00:00Z&sort=ASC&sortBy=createdAt", nil)
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			controller.GetList(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}

func TestGetByID(t *testing.T) {
	refundSvc := serviceMocks.NewIRefundService(t)
	cfg := &config.Config{}
	controller := New(cfg, WithRefundService(refundSvc))

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR: Service returns error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundDetail", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundDetail", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name: "SUCCESS - Submerchant",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				refundSvc.On("GetRefundDetail", constant.ValueCtxMockType(), constant.FilterRefundRequestMockType()).
					Return(nil, nil)
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: uuid.NewString(),
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/refunds/%s", uuid.NewString()), nil)
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			router := chi.NewRouter()
			router.Get("/refunds/{uuid}", controller.GetByID)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}
