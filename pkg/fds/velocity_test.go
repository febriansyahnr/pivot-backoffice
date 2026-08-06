package fds_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/fds"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/stretchr/testify/require"
)

func TestIntegrationVelocityCheckAllow(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := t.Context()

	container, redisExt, err := test.SetupRedis(ctx)
	require.NoError(t, err)
	require.NotNil(t, redisExt)
	require.NotNil(t, container)

	defer container.Terminate(ctx)

	velocityCheck := NewVelocityCheck(redisExt.Client())

	var (
		iter   int
		fdsKey = "fds:velocity-check:payments:merchant-01"
	)

	for range 10 {

		iter++

		rule := VelocityRule{
			Member: fmt.Sprintf("TRX-%04d", iter), // TRX-0001
			Period: time.Minute,
			Rate:   10,
		}

		result, err := velocityCheck.Allow(ctx, fdsKey, rule)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Allowed)
		require.Equal(t, 10-iter, result.Remaining)
		require.Equal(t, 10, result.Limit)
	}

	for range 5 {

		iter++

		rule := VelocityRule{
			Member: fmt.Sprintf("TRX-%04d", iter), // TRX-0011
			Period: time.Minute,
			Rate:   10,
		}

		result, err := velocityCheck.Allow(ctx, fdsKey, rule)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.Allowed)
		require.Equal(t, 0, result.Remaining)
		require.Equal(t, 10, result.Limit)
	}

	require.NoError(t, velocityCheck.Rollback(ctx, fdsKey, "TRX-0010"))
	require.NoError(t, velocityCheck.Rollback(ctx, fdsKey, "TRX-0009"))
	require.NoError(t, velocityCheck.Rollback(ctx, fdsKey, "TRX-0008"))

	rule := VelocityRule{
		Member: "TRX-0011",
		Period: time.Minute,
		Rate:   10,
	}
	result, err := velocityCheck.Allow(ctx, fdsKey, rule)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Allowed)
	require.Equal(t, 2, result.Remaining)
	require.Equal(t, 10, result.Limit)
}
