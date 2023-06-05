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
	zeroCache "github.com/zeromicro/go-zero/core/stores/cache"
	zeroRedis "github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	redisKeyNotExist = errors.New("redis: nil")
	RedisExpiration  = 15 * time.Minute
)

type RedisCache struct {
	clusterClient *redis.ClusterClient
	nodeClient    *redis.Client
	marshal       *marshaler.Marshaler
	expiration    time.Duration
}

func NewRedisCache(cacheRedis zeroCache.CacheConf, expiration time.Duration) Cache {
	var redisInstance *store.RedisStore
	var clusterClient *redis.ClusterClient
	var nodeClient *redis.Client
	if cacheRedis[0].Type == zeroRedis.ClusterType {
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{cacheRedis[0].Host}, Password: cacheRedis[0].Pass})
		redisInstance = store.NewRedis(clusterClient, &store.Options{Expiration: expiration})
	} else {
		nodeClient = redis.NewClient(&redis.Options{Addr: cacheRedis[0].Host, Password: cacheRedis[0].Pass})
		redisInstance = store.NewRedis(nodeClient, &store.Options{Expiration: expiration})
	}

	redisCacheManager := cache.New(redisInstance)
	promMetrics := metrics.NewPrometheus("zkbnb")
	cacheManager := cache.NewMetric(promMetrics, redisCacheManager)
	return &RedisCache{
		clusterClient: clusterClient,
		nodeClient:    nodeClient,
		marshal:       marshaler.New(cacheManager),
		expiration:    expiration,
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

func (c *RedisCache) Close() error {
	if c.nodeClient != nil {
		return c.nodeClient.Close()
	}
	if c.clusterClient != nil {
		return c.clusterClient.Close()
	}
	return nil
}
