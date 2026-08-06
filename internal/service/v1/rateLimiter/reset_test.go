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

func TestRateLimiter_Reset(t *testing.T) {
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
			name: "SUCCESS: Reset Failed Attempts",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectDel(key).SetVal(1)
			},
			wantErr: "",
		},
		{
			name: "ERROR: Error Reset Failed Attempts",
			mockSetup: func(redisMock redismock.ClientMock) {
				redisMock.ExpectDel(key).SetErr(errors.New("error"))
			},
			wantErr: constant.ErrRateLimiterFailedResetAttempts.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(redisClientMock)

			err := service.resetFailedAttempts(ctx, key, lockKey)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.Nil(t, err)
		})
	}

}
