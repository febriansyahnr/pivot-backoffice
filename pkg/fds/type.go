package fds

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type VelocityRule struct {
	Member string
	Period time.Duration
	Rate   int
}

type VelocityResult struct {
	Limit     int
	Allowed   bool
	Remaining int
}

type VelocityChecker interface {
	Allow(ctx context.Context, key string, rule VelocityRule) (*VelocityResult, error)
	Rollback(ctx context.Context, key, member string) error
}

type velocity struct {
	client     *redis.Client
	incrScript *redis.Script
}
