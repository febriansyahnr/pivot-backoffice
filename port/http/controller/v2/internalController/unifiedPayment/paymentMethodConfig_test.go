package v2InternalUnifiedPaymentController_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentMethodConfig(t *testing.T) {
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
				unifiedPaymentSvc.On("GetPaymentMethodConfig", mock.Anything, "123456").
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS: Get payment method config",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetPaymentMethodConfig", mock.Anything, "123456").
					Return(&unifiedPaymentModel.GetPaymentMethodConfigResponse{
						Card: &unifiedPaymentModel.GetPaymentMethodConfigCard{
							Enabled:       true,
							MaximumExpiry: "30 DAYS",
						},
						VirtualAccount: &unifiedPaymentModel.GetPaymentMethodConfigVirtualAccount{
							Enabled: true,
						},
						Qr: &unifiedPaymentModel.GetPaymentMethodConfigQr{
							Enabled: true,
						},
						Ewallet: &unifiedPaymentModel.GetPaymentMethodConfigEWallet{
							Enabled:       true,
							MaximumExpiry: "30 MINUTES",
						},
					}, nil).Once()
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00","data":{"card":{"acceptedChannels":null,"enabled":true,"installmentConfig":null,"maximumAmount":null,"maximumExpiry":"30 DAYS","minimumAmount":null},"ewallet":{"acceptedChannels":null,"enabled":true,"maximumAmount":null,"maximumExpiry":"30 MINUTES","minimumAmount":null},"qr":{"enabled":true,"maximumAmount":null,"maximumExpiry":"","minimumAmount":null},"virtualAccount":{"acceptedChannels":null,"enabled":true,"maximumAmount":null,"maximumExpiry":"","minimumAmount":null}},"message":"Success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/payment-method-config", nil)
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			controller.GetPaymentMethodConfig(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}
