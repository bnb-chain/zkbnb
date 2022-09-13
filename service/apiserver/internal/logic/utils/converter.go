package utils

import (
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func ConvertTx(tx *tx.Tx) *types.Tx {
	return &types.Tx{
		Hash:          tx.TxHash,
		Type:          tx.TxType,
		GasFee:        tx.GasFee,
		GasFeeAssetId: tx.GasFeeAssetId,
		Status:        int64(tx.TxStatus),
		Index:         tx.TxIndex,
		BlockHeight:   tx.BlockHeight,
		NftIndex:      tx.NftIndex,
		PairIndex:     tx.PairIndex,
		CollectionId:  tx.CollectionId,
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
