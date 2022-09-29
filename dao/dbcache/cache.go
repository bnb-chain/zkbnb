package dbcache

import (
	"context"
	"fmt"
)

type QueryFunc func() (interface{}, error)

type Cache interface {
	GetWithSet(ctx context.Context, key string, value interface{}, query QueryFunc) (interface{}, error)
	Get(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Close() error
}

const (
	AccountKeyPrefix = "cache:account_"
	NftKeyPrefix     = "cache:nft_"
	GasAccountKey    = "cache:gasAccount"
	GasConfigKey     = "cache:gasConfig"
)

func AccountKeyByIndex(accountIndex int64) string {
	return AccountKeyPrefix + fmt.Sprintf("%d", accountIndex)
}

func NftKeyByIndex(nftIndex int64) string {
	return NftKeyPrefix + fmt.Sprintf("%d", nftIndex)
}
