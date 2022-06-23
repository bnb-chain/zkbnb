package utils

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
)

func GormTx2Tx(tx *tx.Tx) *types.Tx {
	details := make([]*types.TxDetail, 0)
	if tx.TxDetails != nil {
		for _, detail := range tx.TxDetails {
			d := &types.TxDetail{
				TxId:            detail.TxId,
				AssetId:         detail.AssetId,
				AssetType:       detail.AssetType,
				AccountIndex:    detail.AccountIndex,
				AccountName:     detail.AccountName,
				Balance:         detail.Balance,
				BalanceDelta:    detail.BalanceDelta,
				Order:           detail.Order,
				AccountOrder:    detail.AccountOrder,
				Nonce:           detail.Nonce,
				CollectionNonce: detail.CollectionNonce,
			}
			details = append(details, d)
		}
	}
	return &types.Tx{
		TxHash:        tx.TxHash,
		TxType:        tx.TxType,
		GasFee:        tx.GasFee,
		GasFeeAssetId: tx.GasFeeAssetId,
		TxStatus:      tx.TxStatus,
		BlockHeight:   tx.BlockHeight,
		BlockId:       tx.BlockId,
		StateRoot:     tx.StateRoot,
		NftIndex:      tx.NftIndex,
		PairIndex:     tx.PairIndex,
		AssetId:       tx.AssetId,
		TxAmount:      tx.TxAmount,
		NativeAddress: tx.NativeAddress,
		TxInfo:        tx.TxInfo,
		TxDetails:     details,
		ExtraInfo:     tx.ExtraInfo,
		Memo:          tx.Memo,
		AccountIndex:  tx.AccountIndex,
		Nonce:         tx.Nonce,
		ExpiredAt:     tx.ExpiredAt,
		CreatedAt:     tx.CreatedAt.Unix(),
	}
}

func MempoolTx2Tx(tx *mempool.MempoolTx) *types.Tx {
	details := make([]*types.TxDetail, 0)
	for _, detail := range tx.MempoolDetails {
		d := &types.TxDetail{
			TxId:         detail.TxId,
			AssetId:      detail.AssetId,
			AssetType:    detail.AssetType,
			AccountIndex: detail.AccountIndex,
			AccountName:  detail.AccountName,
			BalanceDelta: detail.BalanceDelta,
			Order:        detail.Order,
			AccountOrder: detail.AccountOrder,
		}
		details = append(details, d)
	}
	return &types.Tx{
		TxHash:        tx.TxHash,
		TxType:        tx.TxType,
		GasFee:        tx.GasFee,
		GasFeeAssetId: tx.GasFeeAssetId,
		TxStatus:      int64(tx.Status),
		BlockHeight:   tx.L2BlockHeight,
		NftIndex:      tx.NftIndex,
		PairIndex:     tx.PairIndex,
		AssetId:       tx.AssetId,
		TxAmount:      tx.TxAmount,
		NativeAddress: tx.NativeAddress,
		TxInfo:        tx.TxInfo,
		TxDetails:     details,
		ExtraInfo:     tx.ExtraInfo,
		Memo:          tx.Memo,
		AccountIndex:  tx.AccountIndex,
		Nonce:         tx.Nonce,
		ExpiredAt:     tx.ExpiredAt,
	}
}
