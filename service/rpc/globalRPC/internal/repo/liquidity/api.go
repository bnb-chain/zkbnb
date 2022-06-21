package liquidity

import (
	table "github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type Liquidity interface {
	GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error)
	GetAllLiquidityAssets() (entity []*table.Liquidity, err error)
}

func New(svcCtx *svc.ServiceContext) Liquidity {
	return &liquidity{
		table: `liquidity`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
