package init

import (
	"github.com/zecrey-labs/zecrey/common/model/l1asset"
	"github.com/zecrey-labs/zecrey/common/model/l2asset"
)

func initAssetsInfo() []*l2asset.L2AssetInfo {
	return []*l2asset.L2AssetInfo{
		// rey
		{
			L2AssetId:   REY_Asset_Id,
			L2AssetName: REY_Asset_Name,
			L2Decimals:  REY_L2_Asset_Decimals,
			L2Symbol:    REY_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      REY_Asset_Id,
					AssetName:    REY_Asset_Name,
					AssetSymbol:  REY_Asset_Name,
					AssetAddress: Ethereum_reyErc20Addr,
					Decimals:     REY_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      REY_Asset_Id,
					AssetName:    REY_Asset_Name,
					AssetSymbol:  REY_Asset_Name,
					AssetAddress: Polygon_reyErc20Addr,
					Decimals:     REY_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      REY_Asset_Id,
					AssetName:    REY_Asset_Name,
					AssetSymbol:  REY_Asset_Name,
					AssetAddress: Aurora_reyErc20Addr,
					Decimals:     REY_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      REY_Asset_Id,
					AssetName:    REY_Asset_Name,
					AssetSymbol:  REY_Asset_Name,
					AssetAddress: Avalanche_reyErc20Addr,
					Decimals:     REY_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      REY_Asset_Id,
					AssetName:    REY_Asset_Name,
					AssetSymbol:  REY_Asset_Name,
					AssetAddress: BSC_reyErc20Addr,
					Decimals:     REY_L1_Asset_Decimals,
				},
			},
		},

		// eth
		{
			L2AssetId:   ETH_Asset_Id,
			L2AssetName: ETH_Asset_Name,
			L2Decimals:  ETH_L2_Asset_Decimals,
			L2Symbol:    ETH_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      ETH_Asset_Id,
					AssetName:    ETH_Asset_Name,
					AssetSymbol:  ETH_Asset_Name,
					AssetAddress: Ethereum_ethErc20Addr,
					Decimals:     ETH_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      ETH_Asset_Id,
					AssetName:    ETH_Asset_Name,
					AssetSymbol:  ETH_Asset_Name,
					AssetAddress: Polygon_ethErc20Addr,
					Decimals:     ETH_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      ETH_Asset_Id,
					AssetName:    ETH_Asset_Name,
					AssetSymbol:  ETH_Asset_Name,
					AssetAddress: Aurora_ethErc20Addr,
					Decimals:     ETH_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      ETH_Asset_Id,
					AssetName:    ETH_Asset_Name,
					AssetSymbol:  ETH_Asset_Name,
					AssetAddress: Avalanche_ethErc20Addr,
					Decimals:     ETH_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      ETH_Asset_Id,
					AssetName:    ETH_Asset_Name,
					AssetSymbol:  ETH_Asset_Name,
					AssetAddress: BSC_ethErc20Addr,
					Decimals:     ETH_L1_Asset_Decimals,
				},
			},
		},

		// matic
		{
			L2AssetId:   MATIC_Asset_Id,
			L2AssetName: MATIC_Asset_Name,
			L2Decimals:  MATIC_L2_Asset_Decimals,
			L2Symbol:    MATIC_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      MATIC_Asset_Id,
					AssetName:    MATIC_Asset_Name,
					AssetSymbol:  MATIC_Asset_Name,
					AssetAddress: Ethereum_maticErc20Addr,
					Decimals:     MATIC_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      MATIC_Asset_Id,
					AssetName:    MATIC_Asset_Name,
					AssetSymbol:  MATIC_Asset_Name,
					AssetAddress: Polygon_maticErc20Addr,
					Decimals:     MATIC_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      MATIC_Asset_Id,
					AssetName:    MATIC_Asset_Name,
					AssetSymbol:  MATIC_Asset_Name,
					AssetAddress: Aurora_maticErc20Addr,
					Decimals:     MATIC_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      MATIC_Asset_Id,
					AssetName:    MATIC_Asset_Name,
					AssetSymbol:  MATIC_Asset_Name,
					AssetAddress: Avalanche_maticErc20Addr,
					Decimals:     MATIC_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      MATIC_Asset_Id,
					AssetName:    MATIC_Asset_Name,
					AssetSymbol:  MATIC_Asset_Name,
					AssetAddress: BSC_maticErc20Addr,
					Decimals:     MATIC_L1_Asset_Decimals,
				},
			},
		},

		// near
		{
			L2AssetId:   NEAR_Asset_Id,
			L2AssetName: NEAR_Asset_Name,
			L2Decimals:  NEAR_L2_Asset_Decimals,
			L2Symbol:    NEAR_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      NEAR_Asset_Id,
					AssetName:    NEAR_Asset_Name,
					AssetSymbol:  NEAR_Asset_Name,
					AssetAddress: Ethereum_nearErc20Addr,
					Decimals:     NEAR_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      NEAR_Asset_Id,
					AssetName:    NEAR_Asset_Name,
					AssetSymbol:  NEAR_Asset_Name,
					AssetAddress: Polygon_nearErc20Addr,
					Decimals:     NEAR_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      NEAR_Asset_Id,
					AssetName:    NEAR_Asset_Name,
					AssetSymbol:  NEAR_Asset_Name,
					AssetAddress: Aurora_nearErc20Addr,
					Decimals:     NEAR_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      NEAR_Asset_Id,
					AssetName:    NEAR_Asset_Name,
					AssetSymbol:  NEAR_Asset_Name,
					AssetAddress: Avalanche_nearErc20Addr,
					Decimals:     NEAR_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      NEAR_Asset_Id,
					AssetName:    NEAR_Asset_Name,
					AssetSymbol:  NEAR_Asset_Name,
					AssetAddress: BSC_nearErc20Addr,
					Decimals:     NEAR_L1_Asset_Decimals,
				},
			},
		},

		// avax
		{
			L2AssetId:   AVAX_Asset_Id,
			L2AssetName: AVAX_Asset_Name,
			L2Decimals:  AVAX_L2_Asset_Decimals,
			L2Symbol:    AVAX_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      AVAX_Asset_Id,
					AssetName:    AVAX_Asset_Name,
					AssetSymbol:  AVAX_Asset_Name,
					AssetAddress: Ethereum_avaxErc20Addr,
					Decimals:     AVAX_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      AVAX_Asset_Id,
					AssetName:    AVAX_Asset_Name,
					AssetSymbol:  AVAX_Asset_Name,
					AssetAddress: Polygon_avaxErc20Addr,
					Decimals:     AVAX_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      AVAX_Asset_Id,
					AssetName:    AVAX_Asset_Name,
					AssetSymbol:  AVAX_Asset_Name,
					AssetAddress: Aurora_avaxErc20Addr,
					Decimals:     AVAX_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      AVAX_Asset_Id,
					AssetName:    AVAX_Asset_Name,
					AssetSymbol:  AVAX_Asset_Name,
					AssetAddress: Avalanche_avaxErc20Addr,
					Decimals:     AVAX_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      AVAX_Asset_Id,
					AssetName:    AVAX_Asset_Name,
					AssetSymbol:  AVAX_Asset_Name,
					AssetAddress: BSC_avaxErc20Addr,
					Decimals:     AVAX_L1_Asset_Decimals,
				},
			},
		},

		// bit
		{
			L2AssetId:   BIT_Asset_Id,
			L2AssetName: BIT_Asset_Name,
			L2Decimals:  BIT_L2_Asset_Decimals,
			L2Symbol:    BIT_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      BIT_Asset_Id,
					AssetName:    BIT_Asset_Name,
					AssetSymbol:  BIT_Asset_Name,
					AssetAddress: Ethereum_bitErc20Addr,
					Decimals:     BIT_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      BIT_Asset_Id,
					AssetName:    BIT_Asset_Name,
					AssetSymbol:  BIT_Asset_Name,
					AssetAddress: Polygon_bitErc20Addr,
					Decimals:     BIT_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      BIT_Asset_Id,
					AssetName:    BIT_Asset_Name,
					AssetSymbol:  BIT_Asset_Name,
					AssetAddress: Aurora_bitErc20Addr,
					Decimals:     BIT_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      BIT_Asset_Id,
					AssetName:    BIT_Asset_Name,
					AssetSymbol:  BIT_Asset_Name,
					AssetAddress: Avalanche_bitErc20Addr,
					Decimals:     BIT_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      BIT_Asset_Id,
					AssetName:    BIT_Asset_Name,
					AssetSymbol:  BIT_Asset_Name,
					AssetAddress: BSC_bitErc20Addr,
					Decimals:     BIT_L1_Asset_Decimals,
				},
			},
		},

		// usdt
		{
			L2AssetId:   USDT_Asset_Id,
			L2AssetName: USDT_Asset_Name,
			L2Decimals:  USDT_L2_Asset_Decimals,
			L2Symbol:    USDT_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      USDT_Asset_Id,
					AssetName:    USDT_Asset_Name,
					AssetSymbol:  USDT_Asset_Name,
					AssetAddress: Ethereum_usdtErc20Addr,
					Decimals:     USDT_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      USDT_Asset_Id,
					AssetName:    USDT_Asset_Name,
					AssetSymbol:  USDT_Asset_Name,
					AssetAddress: Polygon_usdtErc20Addr,
					Decimals:     USDT_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      USDT_Asset_Id,
					AssetName:    USDT_Asset_Name,
					AssetSymbol:  USDT_Asset_Name,
					AssetAddress: Aurora_usdtErc20Addr,
					Decimals:     USDT_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      USDT_Asset_Id,
					AssetName:    USDT_Asset_Name,
					AssetSymbol:  USDT_Asset_Name,
					AssetAddress: Avalanche_usdtErc20Addr,
					Decimals:     USDT_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      USDT_Asset_Id,
					AssetName:    USDT_Asset_Name,
					AssetSymbol:  USDT_Asset_Name,
					AssetAddress: BSC_usdtErc20Addr,
					Decimals:     USDT_L1_Asset_Decimals,
				},
			},
		},

		// usdc
		{
			L2AssetId:   USDC_Asset_Id,
			L2AssetName: USDC_Asset_Name,
			L2Decimals:  USDC_L2_Asset_Decimals,
			L2Symbol:    USDC_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      USDC_Asset_Id,
					AssetName:    USDC_Asset_Name,
					AssetSymbol:  USDC_Asset_Name,
					AssetAddress: Ethereum_usdcErc20Addr,
					Decimals:     USDC_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      USDC_Asset_Id,
					AssetName:    USDC_Asset_Name,
					AssetSymbol:  USDC_Asset_Name,
					AssetAddress: Polygon_usdcErc20Addr,
					Decimals:     USDC_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      USDC_Asset_Id,
					AssetName:    USDC_Asset_Name,
					AssetSymbol:  USDC_Asset_Name,
					AssetAddress: Aurora_usdcErc20Addr,
					Decimals:     USDC_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      USDC_Asset_Id,
					AssetName:    USDC_Asset_Name,
					AssetSymbol:  USDC_Asset_Name,
					AssetAddress: Avalanche_usdcErc20Addr,
					Decimals:     USDC_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      USDC_Asset_Id,
					AssetName:    USDC_Asset_Name,
					AssetSymbol:  USDC_Asset_Name,
					AssetAddress: BSC_usdcErc20Addr,
					Decimals:     USDC_L1_Asset_Decimals,
				},
			},
		},

		// dai
		{
			L2AssetId:   DAI_Asset_Id,
			L2AssetName: DAI_Asset_Name,
			L2Decimals:  DAI_L2_Asset_Decimals,
			L2Symbol:    DAI_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      DAI_Asset_Id,
					AssetName:    DAI_Asset_Name,
					AssetSymbol:  DAI_Asset_Name,
					AssetAddress: Ethereum_daiErc20Addr,
					Decimals:     DAI_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      DAI_Asset_Id,
					AssetName:    DAI_Asset_Name,
					AssetSymbol:  DAI_Asset_Name,
					AssetAddress: Polygon_daiErc20Addr,
					Decimals:     DAI_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      DAI_Asset_Id,
					AssetName:    DAI_Asset_Name,
					AssetSymbol:  DAI_Asset_Name,
					AssetAddress: Aurora_daiErc20Addr,
					Decimals:     DAI_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      DAI_Asset_Id,
					AssetName:    DAI_Asset_Name,
					AssetSymbol:  DAI_Asset_Name,
					AssetAddress: Avalanche_daiErc20Addr,
					Decimals:     DAI_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      DAI_Asset_Id,
					AssetName:    DAI_Asset_Name,
					AssetSymbol:  DAI_Asset_Name,
					AssetAddress: BSC_daiErc20Addr,
					Decimals:     DAI_L1_Asset_Decimals,
				},
			},
		},

		// bnb
		{
			L2AssetId:   BNB_Asset_Id,
			L2AssetName: BNB_Asset_Name,
			L2Decimals:  BNB_L2_Asset_Decimals,
			L2Symbol:    BNB_Asset_Name,
			IsActive:    true,
			L1AssetsInfo: []*l1asset.L1AssetInfo{
				// ethereum
				{
					ChainId:      Eth_Chain_Id,
					AssetId:      BNB_Asset_Id,
					AssetName:    BNB_Asset_Name,
					AssetSymbol:  BNB_Asset_Name,
					AssetAddress: Ethereum_bnbErc20Addr,
					Decimals:     BNB_L1_Asset_Decimals,
				},
				// polygon
				{
					ChainId:      Polygon_Chain_Id,
					AssetId:      BNB_Asset_Id,
					AssetName:    BNB_Asset_Name,
					AssetSymbol:  BNB_Asset_Name,
					AssetAddress: Polygon_bnbErc20Addr,
					Decimals:     BNB_L1_Asset_Decimals,
				},
				// aurora
				{
					ChainId:      Aurora_Chain_Id,
					AssetId:      BNB_Asset_Id,
					AssetName:    BNB_Asset_Name,
					AssetSymbol:  BNB_Asset_Name,
					AssetAddress: Aurora_bnbErc20Addr,
					Decimals:     BNB_L1_Asset_Decimals,
				},
				// avalanche
				{
					ChainId:      Avalanche_Chain_Id,
					AssetId:      BNB_Asset_Id,
					AssetName:    BNB_Asset_Name,
					AssetSymbol:  BNB_Asset_Name,
					AssetAddress: Avalanche_bnbErc20Addr,
					Decimals:     BNB_L1_Asset_Decimals,
				},
				// bsc
				{
					ChainId:      BSC_Chain_Id,
					AssetId:      BNB_Asset_Id,
					AssetName:    BNB_Asset_Name,
					AssetSymbol:  BNB_Asset_Name,
					AssetAddress: BSC_bnbErc20Addr,
					Decimals:     BNB_L1_Asset_Decimals,
				},
			},
		},
	}
}
