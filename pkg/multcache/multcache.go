package multcache

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/marshaler"
	"github.com/eko/gocache/v2/store"
)

type multcache struct {
	marshal *marshaler.Marshaler
}

type QueryFunc func(value ...interface{}) (interface{}, error)

func (m *multcache) GetWithSet(ctx context.Context, key string, value interface{}, timeOut uint32,
	query QueryFunc, args ...interface{}) (interface{}, error) {
	returnObj, err := m.marshal.Get(ctx, key, value)
	if err == nil {
		return returnObj, nil
	}
	if err.Error() == errGoCacheKeyNotExist.Error() || err.Error() == errRedisCacheKeyNotExist.Error() {
		result, err := query(args...)
		if err != nil {
			return nil, err
		}
		return result, m.Set(ctx, key, result, timeOut)
	}
	return nil, err
}

func (m *multcache) Get(ctx context.Context, key string, value interface{}) (interface{}, error) {
	returnObj, err := m.marshal.Get(ctx, key, value)
	if err == nil {
		return returnObj, nil
	}
	return nil, err
}

func (m *multcache) Set(ctx context.Context, key string, value interface{}, timeOut uint32) error {
	return m.marshal.Set(ctx, key, value,
		&store.Options{Expiration: time.Duration(timeOut) * time.Second})
}
