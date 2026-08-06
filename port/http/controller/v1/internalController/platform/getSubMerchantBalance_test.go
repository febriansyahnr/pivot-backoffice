package platformInternalController

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubMerchantBalance(t *testing.T) {

	validRequest := &platform.GetBulkBalanceRequest{
		MerchantID: uuid.NewString(),
		Usecase:    "PAYMENT",
	}

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func(svc *serviceMocks.IPlatformService)
		requestHeader map[string]string
		requestBody   func() []byte
		wantStatus    int
		wantResponse  string
	}{
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: uuid.NewString(),
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func(svc *serviceMocks.IPlatformService) {
				svc.On("GetSubMerchantBalances", constant.ValueCtxMockType(), mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []platform.MerchantBalanceResponse{
							platform.MerchantBalanceResponse{
								MerchantID: uuid.Max.String(),
								AvailableBalance: &platform.PlatformAvailableBalanceResponse{
									Value:    1000,
									Currency: "IDR",
								},
							},
						},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 1,
							TotalPages: 1,
						},
					}, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00","message":"Success","data":[{"merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","availableBalance":{"value":1000,"currency":"IDR"}}],"pagination":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}`,
		},
		{
			name:          "ERROR: Invalid Claims",
			merchantClaim: nil,
			requestBody: func() []byte {
				reqBody, _ := json.Marshal([]byte(`{}`))
				return reqBody
			},
			setupMock: func(svc *serviceMocks.IPlatformService) {
			},
			wantStatus:   http.StatusUnauthorized,
			wantResponse: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Request Payload",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: uuid.NewString(),
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal([]byte(`{}}`))
				return reqBody
			},
			setupMock: func(svc *serviceMocks.IPlatformService) {
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request body"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Field Validation",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: uuid.NewString(),
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(&platform.GetBulkBalanceRequest{
					MerchantID: uuid.NewString(),
					Usecase:    "",
				})
				return reqBody
			},
			setupMock: func(svc *serviceMocks.IPlatformService) {
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'GetBulkBalanceRequest.Usecase' Error:Field validation for 'Usecase' failed on the 'required' tag"}],"traceId":""}}`,
		},
		{
			name: "ERROR: Service Error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: uuid.NewString(),
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func(svc *serviceMocks.IPlatformService) {
				svc.On("GetSubMerchantBalances", constant.ValueCtxMockType(), mock.Anything).
					Return(nil, fmt.Errorf("error"))
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			platformSvc := serviceMocks.NewIPlatformService(t)
			cfg := &config.Config{}
			controller := New(cfg, platformSvc)

			if test.setupMock != nil {
				test.setupMock(platformSvc)
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/balances/sub-merchants", bytes.NewBuffer(reqBody))
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
			router.Post("/balances/sub-merchants", controller.GetSubMerchantBalance)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantResponse, rec.Body.String(), "expected: %s | actual: %s", test.wantResponse, rec.Body.String())
		})
	}
}
