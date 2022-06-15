package init

import asset "github.com/zecrey-labs/zecrey-legend/common/model/assetInfo"

// TODO l2 asset monitor
func initAssetsInfo() []*asset.AssetInfo {
	return []*asset.AssetInfo{
		{
			AssetId:     0,
			L1Address:   "0x00",
			AssetName:   "BNB",
			AssetSymbol: "BNB",
			Decimals:    18,
			Status:      0,
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
