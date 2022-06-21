package liquidity

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

//go:generate mockgen -source api.go -destination api_mock.go -package liquidity

type LiquidityModel interface {
	GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error)
	GetAllLiquidityAssets() (entity []*table.Liquidity, err error)
}

func New(svcCtx *svc.ServiceContext) LiquidityModel {
	return &liquidityModel{
		table: `liquidity`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
