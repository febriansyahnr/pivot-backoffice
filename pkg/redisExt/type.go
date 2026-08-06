package redisExt

import (
	redisRate "github.com/go-redis/redis_rate/v10"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RedisExt")

type Limit = redisRate.Limit
type Result = redisRate.Result
