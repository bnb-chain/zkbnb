package tx

import (
	"gorm.io/gorm"
)

type TxDetailDB struct {
	gorm.Model
	TxId         int64 `gorm:"index"`
	AssetId      int64
	AssetType    int64
	AccountIndex int64 `gorm:"index"`
	AccountName  string
	BalanceEnc   string
	BalanceDelta string
	ChainId      int64
}

type TxDB struct {
	gorm.Model
	TxHash        string `gorm:"uniqueIndex"`
	TxType        int64
	GasFee        int64
	GasFeeAssetId int64
	TxStatus      int64
	BlockHeight   int64 `gorm:"index"`
	BlockId       int64 `gorm:"index"`
	AccountRoot   string
	AssetAId      int64
	AssetBId      int64
	TxAmount      int64
	NativeAddress string
	ChainId       int64
	TxInfo        string
	TxDetails     []*TxDetailDB `gorm:"foreignkey:TxId"`
	ExtraInfo     string
	Memo          string
	// block detail pk
	BlockDetailPk uint
}
