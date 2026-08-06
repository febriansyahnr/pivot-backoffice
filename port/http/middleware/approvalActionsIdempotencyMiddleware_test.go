package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApprovalActionsIdempotencyMiddleware(t *testing.T) {
	t.Skip("need to validate the mock")
	const (
		headerIdempotencyKey = "X-Idempotent-Key"
		testIdempotencyKey   = "test-idempotency-key-123"
		serviceName          = "backend-portal"
	)

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		handlerSetup   func(w http.ResponseWriter, r *http.Request)
		mockSetup      func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger)
		wantStatusCode int
	}{
		{
			name: "SUCCESS: Request without daily limit error - no Redis Del called",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"message":"success"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				// No Redis Del should be called for successful responses
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "SUCCESS: Too many requests with ErrDailyLimitReached - Redis Del called",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":"46","message":"` + constant.ErrDailyLimitReached.Error() + `"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				expectedKey := "backend-portal:pdk-idempotency:POST:approval-actions:" + testIdempotencyKey
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				redisMock.On("Del", mock.AnythingOfType("*context.emptyCtx"), expectedKey).Return(intCmd).Once()
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name: "SUCCESS: Too many requests with ErrInvalidBatchPayoutItem - Redis Del called",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":"99","message":"` + constant.ErrInvalidBatchPayoutItem.Error() + `"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				expectedKey := "backend-portal:pdk-idempotency:POST:approval-actions:" + testIdempotencyKey
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				redisMock.On("Del", mock.AnythingOfType("*context.emptyCtx"), expectedKey).Return(intCmd).Once()
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name: "SUCCESS: Too many requests with HttpStatusErrorDailyLimitReached and remaining message - Redis Del called",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":"` + response.HttpStatusErrorDailyLimitReached + `","message":"Remaining amount today: Rp 1000000"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				expectedKey := "backend-portal:pdk-idempotency:POST:approval-actions:" + testIdempotencyKey
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				redisMock.On("Del", mock.AnythingOfType("*context.emptyCtx"), expectedKey).Return(intCmd).Once()
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name: "ERROR: Redis Del fails - error logged",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":"46","message":"` + constant.ErrDailyLimitReached.Error() + `"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				expectedKey := "backend-portal:pdk-idempotency:POST:approval-actions:" + testIdempotencyKey
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetErr(errors.New("redis connection error"))
				redisMock.On("Del", mock.AnythingOfType("*context.emptyCtx"), expectedKey).Return(intCmd).Once()
				loggerMock.On("Error", mock.AnythingOfType("*context.emptyCtx"), "error when delete idempotency key on approval actions", mock.Anything).Once()
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name: "SUCCESS: Too many requests with other error - no Redis Del called",
			reqSetting: func(req *http.Request) {
				req.Header.Set(headerIdempotencyKey, testIdempotencyKey)
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":"99","message":"rate limit exceeded"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				// No Redis Del should be called for non-daily-limit errors
			},
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name: "SUCCESS: Request without idempotency key header",
			reqSetting: func(req *http.Request) {
				// No idempotency key header set
			},
			handlerSetup: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"message":"success"}`))
			},
			mockSetup: func(redisMock *redisMocks.IRedisExt, loggerMock *loggerMocks.ILogger) {
				// No Redis operations expected
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			redisMock := redisMocks.NewIRedisExt(t)
			loggerMock := loggerMocks.NewILogger(t)

			router := chi.NewRouter()
			router.Use(middleware.ApprovalActionsIdempotencyMiddleware(redisMock, serviceName, loggerMock))

			router.Post("/test", func(w http.ResponseWriter, r *http.Request) {
				test.handlerSetup(w, r)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/test", nil)

			test.reqSetting(req)
			test.mockSetup(redisMock, loggerMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)

			// Verify all mock expectations
			redisMock.AssertExpectations(t)
			loggerMock.AssertExpectations(t)
		})
	}
}

func TestApprovalActionsIdempotencyMiddleware_WithoutResponseWriter(t *testing.T) {
	t.Skip("need to validate the mock")
	const (
		headerIdempotencyKey = "X-Idempotent-Key"
		testIdempotencyKey   = "test-idempotency-key-123"
		serviceName          = "backend-portal"
	)

	redisMock := redisMocks.NewIRedisExt(t)
	loggerMock := loggerMocks.NewILogger(t)

	router := chi.NewRouter()
	// Not using chiExtMiddleware.NewResponseWriterMiddleware
	router.Use(middleware.ApprovalActionsIdempotencyMiddleware(redisMock, serviceName, loggerMock))

	router.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"code":"46","message":"` + constant.ErrDailyLimitReached.Error() + `"}`))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set(headerIdempotencyKey, testIdempotencyKey)

	// No Redis Del should be called because ResponseWriter is not the extended type
	// No mock setup needed

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Result().StatusCode)

	// Verify no unexpected calls were made
	redisMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}
