package multcache

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/marshaler"
	"github.com/eko/gocache/v2/metrics"
	"github.com/eko/gocache/v2/store"
	"github.com/go-redis/redis/v8"
	gocache "github.com/patrickmn/go-cache"
)

// Query function when key does not exist
type MultCache interface {
	GetWithSet(ctx context.Context, key string, value interface{}, timeOut uint32,
		query QueryFunc) error
	Get(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, timeOut uint32) error
}

type Book struct {
	ID   string
	Name string
}

// func NewLoadableCache(expiration, cleanupInterval uint32) MultCache {
// 	gocacheClient := gocache.New(time.Duration(expiration)*time.Minute,
// 		time.Duration(cleanupInterval)*time.Minute)
// 	gocacheStore := store.NewGoCache(gocacheClient, nil)
// 	// goCacheManager := cache.New(gocacheStore)
// 	// promMetrics := metrics.NewPrometheus("my-amazing-app")
// 	loadFunction := func(ctx context.Context, key string) (*Book, error) {
// 		// ... retrieve value from available source
// 		return &Book{ID: "1", Name: "My test amazing book"}, nil
// 	}
// 	cacheManager := cache.NewLoadable(loadFunction, cache.New(gocacheStore))

// 	return &multcache{
// 		marshal: marshaler.New(cacheManager),
// 	}
// }

func NewGoCache(expiration, cleanupInterval uint32) MultCache {
	gocacheClient := gocache.New(time.Duration(expiration)*time.Minute,
		time.Duration(cleanupInterval)*time.Minute)
	gocacheStore := store.NewGoCache(gocacheClient, nil)
	goCacheManager := cache.New(gocacheStore)
	promMetrics := metrics.NewPrometheus("my-amazing-app")
	cacheManager := cache.NewMetric(promMetrics, goCacheManager)
	return &multcache{
		marshal: marshaler.New(cacheManager),
	}
}

func NewRedisCache(redisAdd string, expiration uint32) MultCache {
	redisClient := redis.NewClient(&redis.Options{Addr: redisAdd})
	redisStore := store.NewRedis(redisClient,
		&store.Options{Expiration: time.Duration(expiration) * time.Minute})
	redisCacheManager := cache.New(redisStore)
	promMetrics := metrics.NewPrometheus("my-amazing-app")
	cacheManager := cache.NewMetric(promMetrics, redisCacheManager)
	return &multcache{
		marshal: marshaler.New(cacheManager),
	}
}

func NewMultCache(redisAdd string, expiration, cleanupInterval uint32) MultCache {
	gocacheClient := gocache.New(time.Duration(expiration)*time.Minute,
		time.Duration(cleanupInterval)*time.Minute)
	gocacheStore := store.NewGoCache(gocacheClient, nil)
	goCacheManager := cache.New(gocacheStore)

	redisClient := redis.NewClient(&redis.Options{Addr: redisAdd})
	redisStore := store.NewRedis(redisClient,
		&store.Options{Expiration: time.Duration(expiration) * time.Minute})
	redisCacheManager := cache.New(redisStore)

	promMetrics := metrics.NewPrometheus("my-amazing-app")
	cacheManager := cache.NewMetric(promMetrics, cache.NewChain(
		goCacheManager,
		redisCacheManager),
	)
	return &multcache{
		marshal: marshaler.New(cacheManager),
	}
}
