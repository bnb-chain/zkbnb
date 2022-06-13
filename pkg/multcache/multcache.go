package multcache

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/marshaler"
	"github.com/eko/gocache/v2/store"
)

type multcache struct {
	marshal *marshaler.Marshaler
	timeOut uint32
	ctx     context.Context
}

type QueryFunc func(value ...interface{}) (interface{}, error)

func (m *multcache) GetWithSet(key string, value interface{},
	query QueryFunc, args ...interface{}) (interface{}, error) {
	returnObj, err := m.marshal.Get(m.ctx, key, value)
	if err == nil {
		return returnObj, nil
	}
	if err.Error() == errGoCacheKeyNotExist.Error() || err.Error() == errRedisCacheKeyNotExist.Error() {
		result, err := query(args...)
		if err != nil {
			return nil, err
		}
		return result, m.Set(key, result)
	}
	return nil, err
}

func (m *multcache) Get(key string, value interface{}) (interface{}, error) {
	returnObj, err := m.marshal.Get(m.ctx, key, value)
	if err == nil {
		return returnObj, nil
	}
	return nil, err
}

func (m *multcache) Set(key string, value interface{}) error {
	return m.marshal.Set(m.ctx, key, value,
		&store.Options{Expiration: time.Duration(m.timeOut) * time.Second})
}
