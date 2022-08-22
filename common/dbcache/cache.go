package dbcache

import (
	"context"
	"fmt"
)

type QueryFunc func() (interface{}, error)

type Cache interface {
	GetWithSet(ctx context.Context, key string, value interface{}, query QueryFunc) (interface{}, error)
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}

const (
	AccountKeyPrefix   = "cache:account_"
	LiquidityKeyPrefix = "cache:liquidity_"
	NftKeyPrefix       = "cache:nft_"
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
