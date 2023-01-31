package utils

import (
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func ConvertTx(transaction *tx.Tx) *types.Tx {
	toAccountIndex := int64(-1)

	switch transaction.TxType {
	case types2.TxTypeMintNft:
		txInfo, err := types2.ParseMintNftTxInfo(transaction.TxInfo)
		if err != nil {
			logx.Errorf("parse mintNft tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	case types2.TxTypeTransfer:
		txInfo, err := types2.ParseTransferTxInfo(transaction.TxInfo)
		if err != nil {
			logx.Errorf("parse transfer tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	case types2.TxTypeTransferNft:
		txInfo, err := types2.ParseTransferNftTxInfo(transaction.TxInfo)
		if err != nil {
			logx.Errorf("parse transferNft tx failed: %s", err.Error())
		} else {
			toAccountIndex = txInfo.ToAccountIndex
		}
	}

	// If tx.VerifyAt field has not been set yet,
	// this field is set to zero by default for the front end
	var verifyAt int64 = 0
	if !transaction.VerifyAt.IsZero() {
		verifyAt = transaction.VerifyAt.Unix()
	}

	var status int64 = 0
	if transaction.TxStatus == tx.StatusPending || transaction.TxStatus == tx.StatusExecuted ||
		transaction.TxStatus == tx.StatusPacked || transaction.TxStatus == tx.StatusCommitted {
		status = tx.StatusProcessing
	}

	return &types.Tx{
		Hash:           transaction.TxHash,
		Type:           transaction.TxType,
		GasFee:         transaction.GasFee,
		GasFeeAssetId:  transaction.GasFeeAssetId,
		Status:         status,
		Index:          transaction.TxIndex,
		BlockHeight:    transaction.BlockHeight,
		NftIndex:       transaction.NftIndex,
		CollectionId:   transaction.CollectionId,
		AssetId:        transaction.AssetId,
		Amount:         transaction.TxAmount,
		NativeAddress:  transaction.NativeAddress,
		Info:           transaction.TxInfo,
		ExtraInfo:      transaction.ExtraInfo,
		Memo:           transaction.Memo,
		AccountIndex:   transaction.AccountIndex,
		Nonce:          transaction.Nonce,
		ExpiredAt:      transaction.ExpiredAt,
		CreatedAt:      transaction.CreatedAt.Unix(),
		VerifyAt:       verifyAt,
		ToAccountIndex: toAccountIndex,
	}
}
