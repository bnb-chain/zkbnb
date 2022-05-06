package logic

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func ConvertMempoolTxToTx(mempoolTx *MempoolTx, accountRoot string, currentBlockHeight int64) (tx *Tx) {
	var details []*TxDetail
	for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
		details = append(details, &TxDetail{
			TxId:         mempoolTxDetail.TxId,
			AssetId:      mempoolTxDetail.AssetId,
			AssetType:    mempoolTxDetail.AssetType,
			AccountIndex: mempoolTxDetail.AccountIndex,
			AccountName:  mempoolTxDetail.AccountName,
			Balance:      mempoolTxDetail.Balance,
			BalanceDelta: mempoolTxDetail.BalanceDelta,
		})
	}
	tx = &Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		GasFee:        mempoolTx.GasFee,
		GasFeeAssetId: mempoolTx.GasFeeAssetId,
		TxStatus:      TxStatusPending,
		BlockHeight:   currentBlockHeight,
		AccountRoot:   accountRoot,
		AssetAId:      mempoolTx.AssetAId,
		AssetBId:      mempoolTx.AssetBId,
		TxAmount:      mempoolTx.TxAmount,
		NativeAddress: mempoolTx.NativeAddress,
		TxInfo:        mempoolTx.TxInfo,
		TxDetails:     details,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		Nonce:         mempoolTx.Nonce,
	}
	return tx
}

func AssetToAssetHistory(asset *AccountAsset, l2BlockHeight int64) (assetHistory *AccountAssetHistory) {
	return &AccountAssetHistory{
		AccountIndex:  asset.AccountIndex,
		AssetId:       asset.AssetId,
		Balance:       asset.Balance,
		L2BlockHeight: l2BlockHeight,
	}
}

func LiquidityAssetToLiquidityAssetHistory(liquidityAsset *AccountLiquidity, l2BlockHeight int64) (liquidityAssetHistory *AccountLiquidityHistory) {
	return &AccountLiquidityHistory{
		AccountIndex:  liquidityAsset.AccountIndex,
		PairIndex:     liquidityAsset.PairIndex,
		AssetAId:      liquidityAsset.AssetAId,
		AssetA:        liquidityAsset.AssetA,
		AssetBId:      liquidityAsset.AssetBId,
		AssetB:        liquidityAsset.AssetB,
		LpAmount:      liquidityAsset.LpAmount,
		L2BlockHeight: l2BlockHeight,
	}
}

func NftAssetToNftAssetHistory(asset *L2Nft, l2BlockHeight int64) (assetHistory *L2NftHistory) {
	return &L2NftHistory{
		NftIndex:            asset.NftIndex,
		CreatorAccountIndex: asset.CreatorAccountIndex,
		OwnerAccountIndex:   asset.OwnerAccountIndex,
		AssetId:             asset.AssetId,
		AssetAmount:         asset.AssetAmount,
		NftContentHash:      asset.NftContentHash,
		NftL1TokenId:        asset.NftL1TokenId,
		NftL1Address:        asset.NftL1Address,
		CollectionId:        asset.CollectionId,
		Status:              asset.Status,
		L2BlockHeight:       l2BlockHeight,
	}
}

func AssetHistoryToAsset(assetHistory *AccountAssetHistory) (asset *AccountAsset) {
	return &AccountAsset{
		AccountIndex: assetHistory.AccountIndex,
		AssetId:      assetHistory.AssetId,
		Balance:      assetHistory.Balance,
	}
}

func LiquidityAssetHistoryToLiquidityAsset(liquidityAssetHistory *AccountLiquidityHistory) (liquidityAsset *AccountLiquidity) {
	return &AccountLiquidity{
		AccountIndex: liquidityAssetHistory.AccountIndex,
		PairIndex:    liquidityAssetHistory.PairIndex,
		AssetAId:     liquidityAssetHistory.AssetAId,
		AssetA:       liquidityAssetHistory.AssetA,
		AssetBId:     liquidityAssetHistory.AssetBId,
		AssetB:       liquidityAssetHistory.AssetB,
		LpAmount:     liquidityAssetHistory.LpAmount,
	}
}
