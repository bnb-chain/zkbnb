package dbcache

import (
	"context"
	"errors"
	"time"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/marshaler"
	"github.com/eko/gocache/v2/metrics"
	"github.com/eko/gocache/v2/store"
	"github.com/go-redis/redis/v8"
)

var (
	redisKeyNotExist = errors.New("redis: nil")
)

type RedisCache struct {
	marshal    *marshaler.Marshaler
	expiration time.Duration
}

func NewRedisCache(redisAdd, password string, expiration time.Duration) Cache {
	client := redis.NewClient(&redis.Options{Addr: redisAdd, Password: password})
	redisInstance := store.NewRedis(client, &store.Options{Expiration: expiration})
	redisCacheManager := cache.New(redisInstance)
	promMetrics := metrics.NewPrometheus("zkbas")
	cacheManager := cache.NewMetric(promMetrics, redisCacheManager)
	return &RedisCache{
		marshal:    marshaler.New(cacheManager),
		expiration: expiration,
	}
}

func (c *RedisCache) GetWithSet(ctx context.Context, key string, valueStruct interface{}, query QueryFunc) (interface{}, error) {
	value, err := c.marshal.Get(ctx, key, valueStruct)
	if err == nil {
		return value, nil
	}
	if err.Error() == redisKeyNotExist.Error() {
		value, err = query()
		if err != nil {
			return nil, err
		}
		return value, c.Set(ctx, key, value)
	}
	return nil, err
}

func (c *RedisCache) Get(ctx context.Context, key string, value interface{}) (interface{}, error) {
	object, err := c.marshal.Get(ctx, key, value)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.marshal.Set(ctx, key, value, &store.Options{Expiration: c.expiration})
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.marshal.Delete(ctx, key)
}
