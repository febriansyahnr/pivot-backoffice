package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_UpdateFailedAttempts(t *testing.T) {
	ctx := context.Background()

	redisClient, redisClientMock := redismock.NewClientMock()
	service := New(pdkLoggerMock, redisExt.WrapRedisClient(redisClient, nil), nil)

	var (
		key          = "key-1"
		lockKey      = "lockkey-1"
		currentTime  = time.Now()
		previousTime = currentTime.Add(time.Duration(-FailedAttemptTimeFrameInMinute) * time.Minute)
		member       = redis.Z{
			Member: currentTime.Unix(),
			Score:  float64(currentTime.Unix()),
		}
	)

	tests := []struct {
		name      string
		mockSetup func(redisMock redismock.ClientMock)
		wantErr   string
	}{
		{
			name: "SUCCESS: Update Failed Attempts",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", currentTime.Unix())).SetVal(1)
			},
			wantErr: "",
		},
		{
			name: "SUCCESS: Update Failed Attempts when there are 5 items removed",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(5)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", currentTime.Unix())).SetVal(1)
			},
			wantErr: "",
		},
		{
			name: "SUCCESS: Update Failed Attempts when total count > max failed attempts",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", currentTime.Unix())).SetVal(3)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
			},
			wantErr: "",
		},
		{
			name: "SUCCESS: Update Failed Attempts when fail add item",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetErr(errors.New("error"))
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetVal(1)
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", currentTime.Unix())).SetVal(3)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
			},
			wantErr: "",
		},
		{
			name: "SUCCESS: Update Failed Attempts when fail add & remove item",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetErr(errors.New("error"))
				redisMock.ExpectZRemRangeByScore(key, "0", fmt.Sprintf("%d", previousTime.Unix())).SetErr(errors.New("error"))
				redisMock.ExpectZCount(key, fmt.Sprintf("%d", previousTime.Unix()), fmt.Sprintf("%d", currentTime.Unix())).SetVal(3)
				redisMock.ExpectExpire(key, LockDuration).SetVal(true)
			},
			wantErr: "",
		},
		{
			name: "ERROR: Error on add, remove and count",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetErr(errors.New("error"))
				redisMock.ExpectZRemRangeByScore(key, mock.Anything, mock.Anything).SetErr(errors.New("error"))
				redisMock.ExpectZCount(key, mock.Anything, mock.Anything).SetErr(errors.New("error"))
			},
			wantErr: constant.ErrRateLimiterFailedUpdateAttempts.Error(),
		},
		{
			name: "ERROR: Error on Set Expiry",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, mock.Anything, mock.Anything).SetVal(1)
				redisMock.ExpectZCount(key, mock.Anything, mock.Anything).SetVal(3)
				redisMock.ExpectExpire(key, LockDuration).SetErr(errors.New("error"))
			},
			wantErr: constant.ErrRateLimiterFailedUpdateAttempts.Error(),
		},
		{
			name: "ERROR: Fail on Set Expiry",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZAdd(key, member).SetVal(1)
				redisMock.ExpectZRemRangeByScore(key, mock.Anything, mock.Anything).SetVal(1)
				redisMock.ExpectZCount(key, mock.Anything, mock.Anything).SetVal(3)
				redisMock.ExpectExpire(key, LockDuration).SetVal(false)
			},
			wantErr: constant.ErrRateLimiterFailedUpdateAttempts.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(redisClientMock)

			err := service.updateFailedAttempts(ctx, key, lockKey, currentTime)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.Nil(t, err)
		})
	}

}
