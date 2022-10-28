package utils

import (
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func ConvertTx(tx *tx.Tx) *types.Tx {
	toAccountIndex := int64(-1)

	switch tx.TxType {
	case types2.TxTypeMintNft:
		txInfo, err := types2.ParseMintNftTxInfo(tx.TxInfo)
		if err != nil {
			logx.Errorf("parse mintNft tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	case types2.TxTypeTransfer:
		txInfo, err := types2.ParseTransferTxInfo(tx.TxInfo)
		if err != nil {
			logx.Errorf("parse transfer tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	case types2.TxTypeTransferNft:
		txInfo, err := types2.ParseTransferNftTxInfo(tx.TxInfo)
		if err != nil {
			logx.Errorf("parse transferNft tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	}

	return &types.Tx{
		Hash:           tx.TxHash,
		Type:           tx.TxType,
		GasFee:         tx.GasFee,
		GasFeeAssetId:  tx.GasFeeAssetId,
		Status:         int64(tx.TxStatus),
		Index:          tx.TxIndex,
		BlockHeight:    tx.BlockHeight,
		NftIndex:       tx.NftIndex,
		CollectionId:   tx.CollectionId,
		AssetId:        tx.AssetId,
		Amount:         tx.TxAmount,
		NativeAddress:  tx.NativeAddress,
		Info:           tx.TxInfo,
		ExtraInfo:      tx.ExtraInfo,
		Memo:           tx.Memo,
		AccountIndex:   tx.AccountIndex,
		Nonce:          tx.Nonce,
		ExpiredAt:      tx.ExpiredAt,
		CreatedAt:      tx.CreatedAt.Unix(),
		ToAccountIndex: toAccountIndex,
	}
}
