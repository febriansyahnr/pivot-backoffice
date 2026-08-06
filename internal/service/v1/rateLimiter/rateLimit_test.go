package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitFailedAttempts(t *testing.T) {
	ctx := context.Background()

	redisClient, redisClientMock := redismock.NewClientMock()
	service := New(pdkLoggerMock, redisExt.WrapRedisClient(redisClient, nil), nil)

	var (
		key        = "backend-portal:rate-limiter:test:1"
		lockKey    = "backend-portal:rate-limiter:test:1:lock"
		timestamp  = time.Now()
		successReq = &ratelimiter.RateLimit{
			Attribute:            "1",
			IsCheckResultCorrect: true,
			FeatureName:          "test",
			Timestamp:            timestamp,
		}
		failedReq = &ratelimiter.RateLimit{
			Attribute:            "1",
			IsCheckResultCorrect: false,
			FeatureName:          "test",
			Timestamp:            timestamp,
		}
		previousTime = timestamp.Add(time.Duration(-FailedAttemptTimeFrameInMinute) * time.Minute)
		member       = redis.Z{
			Member: timestamp.Unix(),
			Score:  float64(timestamp.Unix()),
		}
	)

	tests := []struct {
		name      string
		request   *ratelimiter.RateLimit
		mockSetup func(redisMock redismock.ClientMock)
		wantErr   string
	}{
		{
			name:    "Validate: Correct Check RateLimit Attempt and not locked",
			request: successReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectDel(key).SetVal(1)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: "",
		},
		{
			name:    "Validate: Correct Check RateLimit Attempt and locked",
			request: successReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(121)
				redisMock.ExpectDel(key).SetVal(1)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterExceedFailedAttempts.Error(),
		},

		{
			name:    "Validate: No Error When Correct Check RateLimit Attempt and not locked",
			request: successReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectDel(key).SetVal(1)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: "",
		},
		{
			name:    "Validate: Too many attempts Error when Incorrect Check RateLimit Attempt and locked",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(123)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterExceedFailedAttempts.Error(),
		},
		{
			name:    "Validate: Return Error when Error on Validate Attempt ",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetErr(errors.New("error"))
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterFailedValidate.Error(),
		},
		{
			name:    "Update Failed Attempts: Success Add New Failed Attempts",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetVal(1)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: "",
		},
		{
			name:    "Update Failed Attempts: Success Add New Failed Attempts & locked account",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetVal(5)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: "",
		},
		{
			name:    "Update Failed Attempts: Error on Count Failed Attempts",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetErr(errors.New("err"))
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterFailedUpdateAttempts.Error(),
		},
		{
			name:    "Update Failed Attempts: Error Set Expiry Time on Count > Max Failed Attempts",
			request: failedReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ClearExpect()
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetVal(5)
				redisMock.ExpectExpire(key, LockDuration).SetErr(errors.New("error"))
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterFailedUpdateAttempts.Error(),
		},
		{
			name:    "Reset Failed Attempts: Success reset failed attempts",
			request: successReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetVal(5)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
				redisMock.ExpectDel(key).SetVal(1)
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: "",
		},
		{
			name:    "Reset Failed Attempts: Error reset failed attempts",
			request: successReq,
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.MatchExpectationsInOrder(false)
				redisMock.ClearExpect()
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetVal(-1)
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", timestamp.Unix())).SetVal(5)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
				redisMock.ExpectDel(key).SetErr(errors.New("error"))
				redisMock.ExpectDel(lockKey).SetVal(1)
			},
			wantErr: constant.ErrRateLimiterFailedResetAttempts.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(redisClientMock)

			err := service.RateLimitFailedAttempt(ctx, test.request)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.Nil(t, err)
		})
	}

}
