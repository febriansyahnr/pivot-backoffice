package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	ratelimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestMerchantRateLimiterMiddleware(t *testing.T) {
	rateLimiterMock := mockServices.NewIRateLimiter(t)
	handler := MerchantRateLimiterMiddleware(rateLimiterMock, &config.Config{Environment: "test"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	merchantID := "valid-merchant-id"

	tests := []struct {
		name           string
		userClaims     *merchant.MerchantAuthTokenClaims
		setupMock      func()
		expectedStatus int
		expectedHeader map[string]string
	}{
		{
			name:           "when dont have user claim, then should return error",
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "when metadata was valid, then should setup the header",
			userClaims: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				rateLimiterMock.On("ValidateMerchantRateLimit", constant.ValueCtxMockType(), ratelimiterModel.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Path:       "/test-path",
					HTTPMethod: "GET",
				}).Return(&ratelimiterModel.MerchantRateLimitHeaderMetadata{
					RateLimitLimit:     100,
					RateLimitRemaining: 50,
					RateLimitReset:     int64(1609459200),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{},
		},
		{
			name: "when metadata was nil, then should return default rate-limit header",
			userClaims: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				rateLimiterMock.On("ValidateMerchantRateLimit", constant.ValueCtxMockType(), ratelimiterModel.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Path:       "/test-path",
					HTTPMethod: "GET",
				}).Return(nil, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{},
		},
		{
			name: "when metadata was valid but some value was nil, then should return default rate-limit header for nil value",
			userClaims: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				rateLimiterMock.On("ValidateMerchantRateLimit", constant.ValueCtxMockType(), ratelimiterModel.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Path:       "/test-path",
					HTTPMethod: "GET",
				}).Return(&ratelimiterModel.MerchantRateLimitHeaderMetadata{
					RateLimitLimit:     100,
					RateLimitRemaining: 0,
					RateLimitReset:     int64(1609459200),
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{},
		},
		{
			name: "when error ocurred on validation process, then should return error",
			userClaims: &merchant.MerchantAuthTokenClaims{
				MerchantId: merchantID,
			},
			setupMock: func() {
				rateLimiterMock.On("ValidateMerchantRateLimit", constant.ValueCtxMockType(), ratelimiterModel.MerchantRateLimitRequest{
					MerchantID: merchantID,
					Path:       "/test-path",
					HTTPMethod: "GET",
				}).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedHeader: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req, err := http.NewRequest("GET", "/test-path", nil)
			assert.NoError(t, err)

			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tt.userClaims))
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			for key, value := range tt.expectedHeader {
				assert.Equal(t, value, rr.Header().Get(key))
			}
		})
	}
}
