package init

import (
	"github.com/zecrey-labs/zecrey/common/model/liquidityPair"
)

const (
	REY_ETH_Pair_Index    = 0
	ETH_MATIC_Pair_Index  = 1
	ETH_NEAR_Pair_Index   = 2
	NEAR_AVAX_Pair_Index  = 3
	NEAR_BIT_Pair_Index   = 4
	REY_USDT_Pair_Index   = 5
	ETH_USDT_Pair_Index   = 6
	MATIC_USDT_Pair_Index = 7
	NEAR_USDT_Pair_Index  = 8
	AVAX_USDT_Pair_Index  = 9
	BIT_USDT_Pair_Index   = 10
	USDT_USDC_Pair_Index  = 11
	USDT_DAI_Pair_Index   = 12
	USDC_DAI_Pair_Index   = 13
	USDT_BNB_Pair_Index   = 14
	NEAR_BNB_Pair_Index   = 15

	FeeRate      = 30
	TreasuryRate = 10
)

func initLiquidityPair() []*liquidityPair.LiquidityPair {
	return []*liquidityPair.LiquidityPair{
		// rey - eth
		{
			PairIndex:    REY_ETH_Pair_Index,
			AssetAId:     REY_Asset_Id,
			AssetAName:   REY_Asset_Name,
			AssetBId:     ETH_Asset_Id,
			AssetBName:   ETH_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// eth - matic
		{
			PairIndex:    ETH_MATIC_Pair_Index,
			AssetAId:     ETH_Asset_Id,
			AssetAName:   ETH_Asset_Name,
			AssetBId:     MATIC_Asset_Id,
			AssetBName:   MATIC_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// eth - near
		{
			PairIndex:    ETH_NEAR_Pair_Index,
			AssetAId:     ETH_Asset_Id,
			AssetAName:   ETH_Asset_Name,
			AssetBId:     NEAR_Asset_Id,
			AssetBName:   NEAR_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// near - avax
		{
			PairIndex:    NEAR_AVAX_Pair_Index,
			AssetAId:     NEAR_Asset_Id,
			AssetAName:   NEAR_Asset_Name,
			AssetBId:     AVAX_Asset_Id,
			AssetBName:   AVAX_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// near - bit
		{
			PairIndex:    NEAR_BIT_Pair_Index,
			AssetAId:     NEAR_Asset_Id,
			AssetAName:   NEAR_Asset_Name,
			AssetBId:     BIT_Asset_Id,
			AssetBName:   BIT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// rey - usdt
		{
			PairIndex:    REY_USDT_Pair_Index,
			AssetAId:     REY_Asset_Id,
			AssetAName:   REY_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// eth - usdt
		{
			PairIndex:    ETH_USDT_Pair_Index,
			AssetAId:     ETH_Asset_Id,
			AssetAName:   ETH_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// matic - usdt
		{
			PairIndex:    MATIC_USDT_Pair_Index,
			AssetAId:     MATIC_Asset_Id,
			AssetAName:   MATIC_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// near - usdt
		{
			PairIndex:    NEAR_USDT_Pair_Index,
			AssetAId:     NEAR_Asset_Id,
			AssetAName:   NEAR_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// avax - usdt
		{
			PairIndex:    AVAX_USDT_Pair_Index,
			AssetAId:     AVAX_Asset_Id,
			AssetAName:   AVAX_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// bit - usdt
		{
			PairIndex:    BIT_USDT_Pair_Index,
			AssetAId:     BIT_Asset_Id,
			AssetAName:   BIT_Asset_Name,
			AssetBId:     USDT_Asset_Id,
			AssetBName:   USDT_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// usdt - usdc
		{
			PairIndex:    USDT_USDC_Pair_Index,
			AssetAId:     USDT_Asset_Id,
			AssetAName:   USDT_Asset_Name,
			AssetBId:     USDC_Asset_Id,
			AssetBName:   USDC_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// usdt - dai
		{
			PairIndex:    USDT_DAI_Pair_Index,
			AssetAId:     USDT_Asset_Id,
			AssetAName:   USDT_Asset_Name,
			AssetBId:     DAI_Asset_Id,
			AssetBName:   DAI_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// usdc - dai
		{
			PairIndex:    USDC_DAI_Pair_Index,
			AssetAId:     USDC_Asset_Id,
			AssetAName:   USDC_Asset_Name,
			AssetBId:     DAI_Asset_Id,
			AssetBName:   DAI_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// usdt - bnb
		{
			PairIndex:    USDT_BNB_Pair_Index,
			AssetAId:     USDT_Asset_Id,
			AssetAName:   USDT_Asset_Name,
			AssetBId:     BNB_Asset_Id,
			AssetBName:   BNB_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
		// near - bnb
		{
			PairIndex:    NEAR_BNB_Pair_Index,
			AssetAId:     NEAR_Asset_Id,
			AssetAName:   NEAR_Asset_Name,
			AssetBId:     BNB_Asset_Id,
			AssetBName:   BNB_Asset_Name,
			FeeRate:      FeeRate,
			TreasuryRate: TreasuryRate,
		},
	}
}
