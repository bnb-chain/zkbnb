package accountliquidity

import (
	"gorm.io/gorm"
)

type AccountLiquidityInfo struct {
	gorm.Model
	AccountIndex int64 `gorm:"index"`
	PairIndex    int64
	AssetA       int64
	AssetB       int64
	AssetAR      string
	AssetBR      string
	LpEnc        string
}
