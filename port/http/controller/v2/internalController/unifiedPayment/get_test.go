package v2InternalUnifiedPaymentController_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{},
		},
	}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc))

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		requestURL    string
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
				unifiedPaymentSvc.On("GetSessionList", constant.ValueCtxMockType(), constant.PtrGetListFilterRequest()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "ERROR: Invalid page format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestURL:   "/payments?page=invalid&perPage=5&startDate=2025-01-01T00:00:00Z&endDate=2025-01-01T00:00:00Z&sort=ASC&sortBy=createdAt",
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid page format. Use number format instead"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Invalid startDate format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestURL:   "/payments?page=1&perPage=5&startDate=invalid&endDate=2025-01-01T00:00:00Z&sort=ASC&sortBy=createdAt",
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"}],"traceId":""}}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionList", constant.ValueCtxMockType(), constant.PtrGetListFilterRequest()).
					Return(&commonModel.PaginationResponse{}, nil)
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

			requestURL := test.requestURL
			if requestURL == "" {
				requestURL = "/payments?page=1&perPage=5&startDate=2025-01-01T00:00:00Z&endDate=2025-01-01T00:00:00Z&sort=ASC&sortBy=createdAt"
			}
			req := httptest.NewRequest(http.MethodGet, requestURL, nil)
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

func TestGetBinDetailByBinNumber(t *testing.T) {
	service := serviceMocks.NewIUnifiedPaymentService(t)

	handler := New(nil, nil, WithUnifiedPaymentService(service))

	router := chi.NewRouter()
	router.Get("/bin/{binNumber}", handler.GetBinDetailByBinNumber)

	tests := []struct {
		name           string
		binNumber      string
		merchantAuth   *merchant.MerchantAuthTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Merchant auth not found", // NOSONAR
			setupMock:      func() { /* Empty */ },
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:           "ERROR:Invalid BIN number format", // NOSONAR
			binNumber:      "ABC123",
			merchantAuth:   &merchant.MerchantAuthTokenClaims{},
			setupMock:      func() { /* Empty */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"api_validation_error","message":"Invalid BIN format","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid BIN format"}],"traceId":""}}`, // NOSONAR
		},
		{
			name:           "ERROR:Invalid BIN number min length", // NOSONAR
			binNumber:      "12345",
			merchantAuth:   &merchant.MerchantAuthTokenClaims{},
			setupMock:      func() { /* Empty */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"api_validation_error","message":"Invalid BIN format","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid BIN format"}],"traceId":""}}`, // NOSONAR
		},
		{
			name:           "ERROR:Invalid BIN number max length", // NOSONAR
			binNumber:      "123456789",
			merchantAuth:   &merchant.MerchantAuthTokenClaims{},
			setupMock:      func() { /* Empty */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"api_validation_error","message":"Invalid BIN format","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid BIN format"}],"traceId":""}}`, // NOSONAR
		},
		{
			name:         "ERROR:Some error", // NOSONAR
			merchantAuth: &merchant.MerchantAuthTokenClaims{},
			setupMock: func() {
				service.On("GetCardBinDetail", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:         "SUCCESS", // NOSONAR
			merchantAuth: &merchant.MerchantAuthTokenClaims{},
			setupMock: func() {
				service.On("GetCardBinDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.GetBinDetailResponse{
					BIN:       "123456",  // NOSONAR
					CardType:  "CREDIT",  // NOSONAR
					CardLevel: "CLASSIC", // NOSONAR
					Principal: "VISA",    // NOSONAR
					Issuer:    "BCA",     // NOSONAR
					Country:   "ID",      // NOSONAR
					Currency:  "IDR",     // NOSONAR
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"bin":"123456","cardType":"CREDIT","principal":"VISA","cardLevel":"CLASSIC","issuer":"BCA","country":"ID","currency":"IDR"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			ctx := t.Context()
			if test.binNumber == "" {
				test.binNumber = "1234567"
			}
			if test.merchantAuth != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantAuth)
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/bin/"+test.binNumber, nil)

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Response Body:", rec.Body.String())
			}
		})
	}
}
