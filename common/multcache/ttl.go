package multcache

import (
	"time"
)

const (
	AccountTtl = 1000 * time.Millisecond //cache ttl of account

	AssetTtl     = 1000 * time.Millisecond //cache ttl of asset
	AssetListTtl = 2000 * time.Millisecond //cache ttl of asset list

	NftTtl      = 1000 * time.Millisecond //cache ttl of nft
	NftCountTtl = 2000 * time.Millisecond //cache ttl of nft total count
	NftListTtl  = 2000 * time.Millisecond //cache ttl of nft list

	BlockTtl       = 1000 * time.Millisecond //cache ttl of block
	BlockListTtl   = 2000 * time.Millisecond //cache ttl of block list
	BlockHeightTtl = 500 * time.Millisecond  //cache ttl of current block height
	BlockCountTtl  = 2000 * time.Millisecond //cache ttl of block count

	MempoolTxTtl = 500 * time.Millisecond  //cache ttl of mempool tx
	TxTtl        = 2000 * time.Millisecond //cache ttl of tx
	TxCountTtl   = 2000 * time.Millisecond //cache ttl of tx count

	PriceTtl = 500 * time.Millisecond //cache ttl of currency price

	DauTtl = 5000 * time.Millisecond //cache ttl of dau

	SysconfigTtl = 10000 * time.Millisecond //cache ttl of sysConfig
)
