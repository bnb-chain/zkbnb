package multcache

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/marshaler"
	"github.com/eko/gocache/v2/metrics"
	"github.com/eko/gocache/v2/store"
	gocache "github.com/patrickmn/go-cache"
)

// Query function when key does not exist
type MultCache interface {
	GetWithSet(ctx context.Context, key string, value interface{}, timeOut uint32,
		query QueryFunc) (interface{}, error)
	Get(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, timeOut uint32) error
}

type Book struct {
	ID   string
	Name string
}

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
