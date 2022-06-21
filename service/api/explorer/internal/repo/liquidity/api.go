package liquidity

import (
	table "github.com/bnb-chain/zkbas/common/model/liquidity"
)

type Liquidity interface {
	GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error)
	GetAllLiquidityAssets() (entity []*table.Liquidity, err error)
}
