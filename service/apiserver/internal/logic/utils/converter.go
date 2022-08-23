package utils

import (
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func DbTx2Tx(tx *tx.Tx) *types.Tx {
	return &types.Tx{
		Hash:          tx.TxHash,
		Type:          tx.TxType,
		GasFee:        tx.GasFee,
		GasFeeAssetId: tx.GasFeeAssetId,
		Status:        tx.TxStatus,
		Index:         tx.TxIndex,
		BlockHeight:   tx.BlockHeight,
		StateRoot:     tx.StateRoot,
		NftIndex:      tx.NftIndex,
		PairIndex:     tx.PairIndex,
		AssetId:       tx.AssetId,
		Amount:        tx.TxAmount,
		NativeAddress: tx.NativeAddress,
		Info:          tx.TxInfo,
		ExtraInfo:     tx.ExtraInfo,
		Memo:          tx.Memo,
		AccountIndex:  tx.AccountIndex,
		Nonce:         tx.Nonce,
		ExpiredAt:     tx.ExpiredAt,
		CreatedAt:     tx.CreatedAt.Unix(),
	}
}

func DbMempoolTx2Tx(tx *mempool.MempoolTx) *types.Tx {
	return &types.Tx{
		Hash:          tx.TxHash,
		Type:          tx.TxType,
		GasFee:        tx.GasFee,
		GasFeeAssetId: tx.GasFeeAssetId,
		Status:        int64(tx.Status),
		BlockHeight:   tx.L2BlockHeight,
		NftIndex:      tx.NftIndex,
		PairIndex:     tx.PairIndex,
		AssetId:       tx.AssetId,
		Amount:        tx.TxAmount,
		NativeAddress: tx.NativeAddress,
		Info:          tx.TxInfo,
		ExtraInfo:     tx.ExtraInfo,
		Memo:          tx.Memo,
		AccountIndex:  tx.AccountIndex,
		Nonce:         tx.Nonce,
		ExpiredAt:     tx.ExpiredAt,
	}
}
