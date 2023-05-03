package utils

import (
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func ConvertTx(tx *tx.Tx, cache *cache.MemCache) *types.Tx {
	fromAccountAddress := ""
	toAccountAddress := ""

	switch tx.TxType {
	case types2.TxTypeDeposit:
		fromAccountAddress = tx.NativeAddress
	case types2.TxTypeDepositNft:
		fromAccountAddress = tx.NativeAddress
	case types2.TxTypeWithdraw:
		toAccountAddress = tx.NativeAddress
	case types2.TxTypeWithdrawNft:
		toAccountAddress = tx.NativeAddress
	case types2.TxTypeFullExit:
		toAccountAddress = tx.NativeAddress
	case types2.TxTypeFullExitNft:
		toAccountAddress = tx.NativeAddress
	}

	// If tx.VerifyAt field has not been set yet,
	// this field is set to zero by default for the front end
	var verifyAt int64 = 0
	if !tx.VerifyAt.IsZero() {
		verifyAt = tx.VerifyAt.Unix()
	}

	result := &types.Tx{
		Hash:             tx.TxHash,
		Type:             tx.TxType,
		GasFee:           tx.GasFee,
		GasFeeAssetId:    tx.GasFeeAssetId,
		Status:           int64(tx.TxStatus),
		Index:            tx.TxIndex,
		BlockHeight:      tx.BlockHeight,
		NftIndex:         tx.NftIndex,
		CollectionId:     tx.CollectionId,
		AssetId:          tx.AssetId,
		Amount:           tx.TxAmount,
		NativeAddress:    tx.NativeAddress,
		Info:             tx.TxInfo,
		ExtraInfo:        tx.ExtraInfo,
		Memo:             tx.Memo,
		AccountIndex:     tx.AccountIndex,
		Nonce:            tx.Nonce,
		ExpiredAt:        tx.ExpiredAt,
		CreatedAt:        tx.CreatedAt.Unix(),
		VerifyAt:         verifyAt,
		FromAccountIndex: tx.FromAccountIndex,
		FromL1Address:    fromAccountAddress,
		ToAccountIndex:   tx.ToAccountIndex,
		ToL1Address:      toAccountAddress,
	}
	if tx.AccountIndex >= 0 {
		result.L1Address, _ = cache.GetL1AddressByIndex(tx.AccountIndex)
	}
	result.AssetName, _ = cache.GetAssetNameById(tx.AssetId)
	if tx.FromAccountIndex >= 0 && len(result.FromL1Address) == 0 {
		result.FromL1Address, _ = cache.GetL1AddressByIndex(tx.FromAccountIndex)
	}
	if tx.ToAccountIndex >= 0 && len(result.ToL1Address) == 0 {
		result.ToL1Address, _ = cache.GetL1AddressByIndex(tx.ToAccountIndex)
	}
	return result
}
