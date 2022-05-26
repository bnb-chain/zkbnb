package l2asset

import (
	"gorm.io/gorm"
)

type L2AssetInfo struct {
	gorm.Model
	L2AssetId    int64 `gorm:"uniqueIndex"`
	L2AssetName  string
	L2Decimals   int64
	L2Symbol     string
	IsActive     bool
	L1AssetsInfo []*L1AssetInfo `gorm:"foreignkey:L2AssetPk"`
}

type L1AssetInfo struct {
	gorm.Model
	ChainId           int64 `gorm:"index"`
	AssetId           int64 `gorm:"index"`
	L2AssetPk         int64 `gorm:"index"`
	AssetName         string
	AssetSymbol       string
	AssetAddress      string
	Decimals          int64
	LockedAssetAmount int64
	WithdrawFeeRate   int64
}

type Asset struct {
	AssetId    uint32
	BalanceEnc string
}

type LatestAccountInfo struct {
	AccountIndex uint32
	AccountName  string
	AccountPk    string
	Assets       []*Asset
}
