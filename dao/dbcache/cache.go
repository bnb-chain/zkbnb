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
}

const (
	AccountKeyPrefix   = "cache:account_"
	LiquidityKeyPrefix = "cache:liquidity_"
	NftKeyPrefix       = "cache:nft_"
	GasAccountKey      = "cache:gasAccount"
	GasAssetsKey       = "cache:gasAssets"
)

func AccountKeyByIndex(accountIndex int64) string {
	return AccountKeyPrefix + fmt.Sprintf("%d", accountIndex)
}

func LiquidityKeyByIndex(pairIndex int64) string {
	return LiquidityKeyPrefix + fmt.Sprintf("%d", pairIndex)
}

func NftKeyByIndex(nftIndex int64) string {
	return NftKeyPrefix + fmt.Sprintf("%d", nftIndex)
}

type DummyCache struct{}

func NewDummyCache() *DummyCache {
	return &DummyCache{}
}
func (_ *DummyCache) GetWithSet(_ context.Context, _ string, _ interface{}, query QueryFunc) (interface{}, error) {
	return query()
}
func (_ *DummyCache) Get(_ context.Context, _ string, _ interface{}) (interface{}, error) {
	return nil, fmt.Errorf("not implement")
}
func (_ *DummyCache) Set(_ context.Context, _ string, _ interface{}) error {
	return nil
}
func (_ *DummyCache) Delete(_ context.Context, _ string) error {
	return nil
}
