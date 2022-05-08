package init

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
)

func initAssetsInfo() []*l2asset.L2AssetInfo {
	return []*l2asset.L2AssetInfo{
		{
			AssetId:     0,
			AssetName:   "BNB",
			AssetSymbol: "BNB",
			Decimals:    18,
			Status:      0,
		},
	}
}
