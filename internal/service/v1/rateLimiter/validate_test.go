package ratelimiter

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_ValidateAttempt(t *testing.T) {
	ctx := context.Background()

	redisClient, redisClientMock := redismock.NewClientMock()
	service := New(pdkLoggerMock, redisExt.WrapRedisClient(redisClient, nil), nil)

	var (
		key     = "key-1"
		lockKey = "lockkey-1"
	)

	tests := []struct {
		name      string
		mockSetup func(redisMock redismock.ClientMock)
		wantErr   string
	}{
		{
			name: "SUCCESS: User not locked",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectTTL(key).SetVal(-1)
			},
			wantErr: "",
		},
		{
			name: "ERROR: User locked",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectTTL(key).SetVal(100)
			},
			wantErr: constant.ErrRateLimiterExceedFailedAttempts.Error(),
		},
		{
			name: "ERROR: Get Key Expiry Time",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectSetNX(lockKey, true, ExclusiveLockDuration).SetVal(true)
				redisMock.ExpectTTL(key).SetErr(errors.New("errors"))
			},
			wantErr: constant.ErrRateLimiterFailedValidate.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(redisClientMock)

			err := service.validateAttempt(ctx, key, lockKey)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.Nil(t, err)
		})
	}
}
