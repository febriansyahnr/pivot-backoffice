package v2InternalUnifiedPaymentController_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetChargeList(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{},
		},
	}

	merchantID := "0044e2e0-dd49-45f5-9fdc-66530d5b3a49"
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc))

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		params        string
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
			name:   "ERROR: Invalid page format",
			params: "page=invalid&perPage=5",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid page format. Use number format instead"}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Invalid perPage format",
			params: "page=1&perPage=invalid",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid perPage format. Use number format instead"}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Invalid startDate format",
			params: "page=1&perPage=5&startDate=invalid-date&endDate=2025-01-01T00:00:00Z",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Invalid endDate format",
			params: "page=1&perPage=5&startDate=2025-01-01T00:00:00Z&endDate=invalid-date",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Invalid date range",
			params: "page=1&perPage=5&startDate=2025-01-01T00:00:00Z",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"startDate or endDate cannot be empty"}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Backdate exceeded the limit",
			params: "page=1&perPage=5&startDate=2025-01-01T00:00:00Z&endDate=2025-01-01T00:00:00Z",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months."}],"traceId":""}}`,
		},
		{
			name:   "ERROR: Invalid sort value",
			params: "page=1&perPage=5&sort=XXX",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"sort parameter must be either ASC or DESC"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Service returns error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"GetChargeList", constant.ValueCtxMockType(), constant.PtrFilterChargeRequest(),
				).Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"GetChargeList", constant.ValueCtxMockType(), constant.PtrFilterChargeRequest(),
				).Return(&commonModel.PaginationResponse{}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success", "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
		{
			name: "SUCCESS: With sub-merchant ID",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "parent-merchant-123",
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: "sub-merchant-456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On(
					"GetChargeList", constant.ValueCtxMockType(), constant.PtrFilterChargeRequest(),
				).Return(&commonModel.PaginationResponse{}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success", "pagination":{"page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			url := "/charges"
			if test.params != "" {
				url += "?" + test.params
			} else {
				url += "?page=1&perPage=5&sort=ASC&sortBy=createdAt"
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			ctx := t.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			controller.GetChargeList(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}

func TestGetChargeByID(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{},
		},
	}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc))

	tests := []struct {
		name          string
		chargeID      string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			chargeID:     uuid.NewString(),
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name:     "ERROR: Invalid UUID",
			chargeID: "invalid-uuid-format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:     "ERROR: Service returns error",
			chargeID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetChargeDetail", constant.ValueCtxMockType(), constant.PtrGetUnifiedPaymentChargeRequest()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:     "SUCCESS",
			chargeID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetChargeDetail", constant.ValueCtxMockType(), constant.PtrGetUnifiedPaymentChargeRequest()).
					Return(&unifiedPaymentModel.ChargeResponse{
						ID:        "valid-charge-id",
						ExpiredAt: util.ValueToPtr(time.Now()),
					}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00","message":"Success","data":{"id":"valid-charge-id","paymentSessionId":"","paymentSessionClientReferenceId":"","amount":{"value":0,"currency":""},"statementDescriptor":"","status":"","authorizedAmount":null,"capturedAmount":null,"isCaptured":false,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","paidAt":null}}`,
		},
		{
			name:     "SUCCESS: With sub-merchant ID",
			chargeID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "parent-merchant-123",
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: "sub-merchant-456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetChargeDetail", constant.ValueCtxMockType(), constant.PtrGetUnifiedPaymentChargeRequest()).
					Return(&unifiedPaymentModel.ChargeResponse{
						ID: "valid-charge-id",
					}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00","message":"Success","data":{"id":"valid-charge-id","paymentSessionId":"","paymentSessionClientReferenceId":"","amount":{"value":0,"currency":""},"statementDescriptor":"","status":"","authorizedAmount":null,"capturedAmount":null,"isCaptured":false,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","paidAt":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/charges/%s", test.chargeID), nil)
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
			router.Get("/charges/{uuid}", controller.GetChargeByID)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}
