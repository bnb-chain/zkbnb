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
	AccountKeyPrefix         = "cache:account_"
	NftKeyPrefix             = "cache:nft_"
	GasAccountKey            = "cache:gasAccount"
	GasConfigKey             = "cache:gasConfig"
	AccountNonceKeyPrefix    = "cache:accountNonce_"
	RetryPendingPoolTxPrefix = "cache:retryPendingPoolTx_"
	ChangePubKeyPKPrefix     = "cache:changePubKeyPK_"
)

func AccountKeyByIndex(accountIndex int64) string {
	return AccountKeyPrefix + fmt.Sprintf("%d", accountIndex)
}

func AccountKeyByPK(pk string) string {
	return ChangePubKeyPKPrefix + fmt.Sprintf("%s", pk)
}

func AccountKeyByL1Address(l1Address string) string {
	return AccountKeyPrefix + fmt.Sprintf("%s", l1Address)
}

func AccountNonceKeyByIndex(accountIndex int64) string {
	return AccountNonceKeyPrefix + fmt.Sprintf("%d", accountIndex)
}

func NftKeyByIndex(nftIndex int64) string {
	return NftKeyPrefix + fmt.Sprintf("%d", nftIndex)
}

func PendingPoolTxKeyByPoolTxId(poolTxId uint) string {
	return RetryPendingPoolTxPrefix + fmt.Sprintf("%d", poolTxId)
}
