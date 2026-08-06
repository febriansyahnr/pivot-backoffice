package redisExt

import (
	"context"
	"encoding/json"
)

func ScanHashField(ctx context.Context, redis IRedisExt, key, field string, target interface{}) error {
	result, err := redis.HGet(ctx, key, field).Result()
	if err != nil {
		return err
	}

	if unmarshaler, ok := target.(interface{ UnmarshalBinary([]byte) error }); ok {
		return unmarshaler.UnmarshalBinary([]byte(result))
	}

	return json.Unmarshal([]byte(result), target)
}
