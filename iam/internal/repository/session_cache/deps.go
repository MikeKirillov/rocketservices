package sessioncache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient interface {
	HSet(ctx context.Context, key string, values ...any) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}
