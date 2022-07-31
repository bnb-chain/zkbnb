package multcache

const (
	AccountTtl = 1 //cache ttl of account in second

	AssetTtl     = 1 //cache ttl of asset in second
	AssetListTtl = 2 //cache ttl of asset list in second

	NftTtl      = 1 //cache ttl of nft in second
	NftCountTtl = 2 //cache ttl of nft total count in second
	NftListTtl  = 2 //cache ttl of nft list in second

	BlockTtl       = 2 //cache ttl of block in second
	BlockListTtl   = 2 //cache ttl of block list in second
	BlockHeightTtl = 1 //cache ttl of current block height in second
	BlockCountTtl  = 2 //cache ttl of block count in second

	MempoolTxTtl = 1 //cache ttl of mempool tx in second
	TxTtl        = 2 //cache ttl of tx in second
	TxCountTtl   = 2 //cache ttl of tx count in second

	PriceTtl = 1 //cache ttl of currency price in second

	DauTtl = 5 //cache ttl of dau in second

	SysconfigTtl = 10 //cache ttl of sysconfig in second
)
