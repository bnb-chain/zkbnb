package logic

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
)

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
		ExpiredAt:     mempoolTx.ExpiredAt,
	}
	return tx
}
