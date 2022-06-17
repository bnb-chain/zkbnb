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

func (m *multcache) GetWithSet(ctx context.Context, key string, value interface{}, timeOut uint32,
	query QueryFunc) error {
	_, err := m.marshal.Get(ctx, key, value)
	if err == nil {
		return nil
	}
	if err.Error() == errGoCacheKeyNotExist.Error() || err.Error() == errRedisCacheKeyNotExist.Error() {
		result, err := query()
		if err != nil {
			return err
		}
		err = m.Set(ctx, key, result, timeOut)
		return err
	}
	return err
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
