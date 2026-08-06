package redisExt

import (
	"context"

	"github.com/go-redsync/redsync/v4"
	"go.opentelemetry.io/otel/attribute"
)

type redisMutex struct {
	mutex *redsync.Mutex
}

func (r *redisExt) NewMutex(name string, options ...redsync.Option) IMutexer {
	return &redisMutex{r.rs.NewMutex(name, options...)}
}

func (r *redisMutex) LockContext(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "pkg/redisExt/mutex/Lock")
	defer segment.End()

	segment.SetAttributes(attribute.String("lock.key", r.mutex.Name()))

	return r.mutex.LockContext(ctx)
}

func (r *redisMutex) UnlockContext(ctx context.Context) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "pkg/redisExt/mutex/Unlock")
	defer segment.End()

	segment.SetAttributes(attribute.String("lock.key", r.mutex.Name()))

	return r.mutex.UnlockContext(ctx)
}

func (r *redisMutex) Value() string {
	return r.mutex.Value()
}
