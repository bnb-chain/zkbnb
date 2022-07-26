package l2asset

import (
	"context"
	table "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type L2asset interface {
	GetL2AssetsList(ctx context.Context) (res []*table.AssetInfo, err error)
	GetL2AssetInfoBySymbol(ctx context.Context, symbol string) (res *table.AssetInfo, err error)
	GetSimpleL2AssetInfoByAssetId(ctx context.Context, assetId uint32) (res *table.AssetInfo, err error)
}

func New(svcCtx *svc.ServiceContext) L2asset {
	return &l2asset{
		table: `asset_info`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
