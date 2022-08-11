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

type QueryFunc func() (interface{}, error)

func (m *multcache) GetWithSet(ctx context.Context, key string, valueStruct interface{}, duration time.Duration,
	query QueryFunc) (interface{}, error) {
	value, err := m.marshal.Get(ctx, key, valueStruct)
	if err == nil {
		return value, nil
	}
	if err.Error() == errGoCacheKeyNotExist.Error() || err.Error() == errRedisCacheKeyNotExist.Error() {
		value, err = query()
		if err != nil {
			return nil, err
		}
		return value, m.Set(ctx, key, value, duration)
	}
	return nil, err
}

func (m *multcache) Get(ctx context.Context, key string, value interface{}) (interface{}, error) {
	returnObj, err := m.marshal.Get(ctx, key, value)
	if err != nil {
		return nil, err
	}
	return returnObj, nil
}

func (m *multcache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if err := m.marshal.Set(ctx, key, value, &store.Options{Expiration: duration}); err != nil {
		return err
	}
	return nil
}

func (m *multcache) Delete(ctx context.Context, key string) error {
	if err := m.marshal.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}
