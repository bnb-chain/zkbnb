package l2asset

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
)

type L2asset interface {
	GetL2AssetsList() (res []*table.AssetInfo, err error)
	GetL2AssetInfoBySymbol(symbol string) (res *table.AssetInfo, err error)
	GetSimpleL2AssetInfoByAssetId(assetId uint32) (res *table.AssetInfo, err error)
}

func New(svcCtx *svc.ServiceContext) L2asset {
	return &l2asset{
		table: `l2_asset_info`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}