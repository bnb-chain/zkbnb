package logic

import "github.com/zecrey-labs/zecrey-legend/common/model/tx"

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func ConvertMempoolTxToTx(mempoolTx *MempoolTx, txDetails []*tx.TxDetail, accountRoot string, currentBlockHeight int64) (tx *Tx) {
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
		TxDetails:     txDetails,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		AccountIndex:  mempoolTx.AccountIndex,
		Nonce:         mempoolTx.Nonce,
	}
	return tx
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
