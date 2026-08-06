package fds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewVelocityCheck(client *redis.Client) VelocityChecker {
	incrScript := `
-- KEYS[1]: Redis Key (sorted set)
-- ARGV[1]: Current Timestamp (nanoseconds)
-- ARGV[2]: Window Start Timestamp (nanoseconds)
-- ARGV[3]: Threshold Count
-- ARGV[4]: Member
-- ARGV[5]: Expiry Duration (seconds)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local threshold = tonumber(ARGV[3])
local member = ARGV[4]
local expiry = tonumber(ARGV[5])

-- Step 1: Remove expired entries (cleanup)
redis.call('ZREMRANGEBYSCORE', key, '0', window_start)

-- Step 2: Count current entries in window
local current_count = redis.call('ZCOUNT', key, window_start, '+inf')

-- Step 3: Check if threshold exceeded
if current_count >= threshold then
    return {0, 0, threshold}
end

-- Step 4: If allowed, increment (add to sorted set)
redis.call('ZADD', key, now, member)

-- Step 5: Set expiry
redis.call('EXPIRE', key, expiry)

-- Step 6: Get remaining
local now_count = redis.call('ZCOUNT', key, window_start, '+inf')

return {1, (threshold-now_count), threshold}
`
	return &velocity{
		client:     client,
		incrScript: redis.NewScript(incrScript),
	}
}

func (v *velocity) Allow(ctx context.Context, key string, rule VelocityRule) (*VelocityResult, error) {
	now := time.Now().UTC()
	windowStart := now.Add(-rule.Period)

	// Execute Lua script
	result, err := v.incrScript.Run(
		ctx,
		v.client,
		[]string{key},
		now.UnixNano(),               // ARGV[1]
		windowStart.UnixNano(),       // ARGV[2]
		rule.Rate,                    // ARGV[3]
		rule.Member,                  // ARGV[4]
		int(rule.Period.Seconds())*2, // ARGV[5]. Note: Buffer to prevent premature deletion
	).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to run increment script: %s", err)
	}

	resultSlice, ok := result.([]any)
	if !ok {
		return nil, errors.New("unexpected result format")

	} else if len(resultSlice) != 3 {
		return nil, fmt.Errorf("unexpected result length: expected 3, got %d", len(resultSlice))
	}

	var (
		allowedInt, _ = resultSlice[0].(int64)
		remaining, _  = resultSlice[1].(int64)
		limit, _      = resultSlice[2].(int64)
	)
	return &VelocityResult{
		Limit:     int(limit),
		Allowed:   allowedInt == 1,
		Remaining: int(remaining),
	}, nil
}

func (v *velocity) Rollback(ctx context.Context, key, member string) error {
	return v.client.ZRem(ctx, key, member).Err()
}
