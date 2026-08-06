package redisExt

import (
	"context"
	"fmt"
	"strings"
	"time"

	pdkRedis "github.com/paper-indonesia/pdk/v2/redisExt"

	redisRate "github.com/go-redis/redis_rate/v10"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
)

type IRedisExt interface {
	Client() *redis.Client
	Close() error
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd

	CustomIncr(ctx context.Context, key string, expiration time.Duration) *redis.StringCmd

	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGetScan(ctx context.Context, key, field string, dst interface{}) error
	HGetAllScan(ctx context.Context, key string, dst interface{}) error
	HIncrByFloat(ctx context.Context, key, field string, incr float64) (result float64, err error)
	HIncrBy(ctx context.Context, key, field string, incr int64) (result int64, err error)
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd

	// Redsync
	NewMutex(name string, options ...redsync.Option) IMutexer
}

type ILimiter interface {
	Reset(ctx context.Context, key string) error

	Allow(ctx context.Context, key string, limit *Limit) (*Result, error)
}

type IMutexer interface {
	LockContext(ctx context.Context) error
	UnlockContext(ctx context.Context) (bool, error)
	Value() string
}

type redisExt struct {
	client *redis.Client
	rs     *redsync.Redsync
}

type limiter struct {
	client *redisRate.Limiter
}

var ErrNil error = redis.Nil

func WrapRedisClient(client *redis.Client, rs *redsync.Redsync) IRedisExt {
	return &redisExt{client, rs}
}

func New(config pdkRedis.Config, opts ...pdkRedis.OptionFunc) (IRedisExt, error) {
	redis, err := pdkRedis.New(config, opts...)
	if err != nil {
		return nil, err
	}

	return &redisExt{redis.Client, redis.Redsync}, nil
}

func NewLimiter(db *redis.Client) ILimiter {
	return &limiter{redisRate.NewLimiter(db)}
}

func (r *redisExt) Client() *redis.Client {
	return r.client
}

func (r *redisExt) Close() error {
	return r.client.Close()
}

func (r *redisExt) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Del(ctx, keys...)
}

func (r *redisExt) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

func (r *redisExt) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

func (r *redisExt) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return r.client.TTL(ctx, key)
}

func (r *redisExt) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.client.SetNX(ctx, key, value, expiration)
}

func (r *redisExt) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return r.client.Scan(ctx, cursor, match, count)
}

func (r *redisExt) Incr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Incr(ctx, key)
}

func (r *redisExt) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return r.client.IncrBy(ctx, key, value)
}

// CustomIncr is a custom increment function that will increment the key by 1, if the key is not exist, it will set the key with value 1
func (r *redisExt) CustomIncr(ctx context.Context, key string, expiration time.Duration) *redis.StringCmd {
	data := r.client.Get(ctx, key)
	if data.Val() == "" {
		r.client.Set(ctx, key, 1, expiration)
	} else {
		r.client.Incr(ctx, key)
	}

	return data
}

func (r *redisExt) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}

func (r *redisExt) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return r.client.HGet(ctx, key, field)
}

func (r *redisExt) HGetScan(ctx context.Context, key, field string, dst interface{}) error {
	return r.client.HGet(ctx, key, field).Scan(dst)
}

func (r *redisExt) HGetAllScan(ctx context.Context, key string, dst interface{}) error {

	if err := r.client.HGetAll(ctx, key).Scan(dst); err == nil {
		return nil

	} else if !strings.HasPrefix(err.Error(), "redis.Scan(non-struct") {
		return err
	}

	maps, ok := dst.(*map[string]string)
	if !ok {
		return fmt.Errorf("redis.Scan(non-struct & non-map[string]string %T)", dst)
	}
	*maps = r.client.HGetAll(ctx, key).Val()

	return nil
}

func (r *redisExt) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

func (r *redisExt) HIncrByFloat(ctx context.Context, key, field string, incr float64) (result float64, err error) {
	return r.client.HIncrByFloat(ctx, key, field, incr).Result()
}

func (r *redisExt) HIncrBy(ctx context.Context, key, field string, incr int64) (result int64, err error) {
	return r.client.HIncrBy(ctx, key, field, incr).Result()
}

func (r *redisExt) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

func (r *redisExt) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	return r.client.Keys(ctx, pattern)
}

func (r *redisExt) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Exists(ctx, keys...)
}

func (l *limiter) Reset(ctx context.Context, key string) error {
	return l.client.Reset(ctx, key)
}

func (l *limiter) Allow(ctx context.Context, key string, limit *Limit) (*Result, error) {
	return l.client.Allow(ctx, key, *limit)
}
