package cache

import (
	"context"
)

type QueryFunc func() (interface{}, error)

type Cache interface {
	GetWithSet(ctx context.Context, key string, value interface{}, query QueryFunc) (interface{}, error)
	Get(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}
