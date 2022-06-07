package init

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
)

// TODO l2 asset monitor
func initAssetsInfo() []*l2asset.L2AssetInfo {
	return []*l2asset.L2AssetInfo{
		{
			AssetId:      0,
			AssetAddress: "0x00",
			AssetName:    "BNB",
			AssetSymbol:  "BNB",
			Decimals:     18,
			Status:       0,
		},
		//{
		//	AssetId:     1,
		//	AssetName:   "LEG",
		//	AssetSymbol: "LEG",
		//	Decimals:    18,
		//	Status:      0,
		//},
		//{
		//	AssetId:     2,
		//	AssetName:   "REY",
		//	AssetSymbol: "REY",
		//	Decimals:    18,
		//	Status:      0,
		//},
	}
}
