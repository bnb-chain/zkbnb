package core

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type TransferExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *legendTxTypes.TransferTxInfo
}

func NewTransferExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &TransferExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *TransferExecutor) Prepare() error {
	txInfo, err := commonTx.ParseTransferTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	e.txInfo = txInfo
	return nil
}

func (e *TransferExecutor) verifyExpiredAt() error {
	if !e.bc.dryRun {
		if e.txInfo.ExpiredAt != commonConstant.NilExpiredAt && e.txInfo.ExpiredAt < e.bc.currentBlock.CreatedAt.UnixMilli() {
			return errors.New("invalid ExpiredAt")
		}
	} else {
		if e.txInfo.ExpiredAt < time.Now().UnixMilli() {
			return errors.New("invalid ExpiredAt")
		}
	}
	return nil
}

func (e *TransferExecutor) verifyNonce(accountIndex int64) error {
	if !e.bc.dryRun {
		if e.txInfo.Nonce != e.bc.accountMap[accountIndex].Nonce {
			return errors.New("invalid Nonce")
		}
	} else {
		nonce, err := e.bc.getPendingNonce(accountIndex)
		if err != nil {
			return errors.New("cannot verify nonce")
		}
		if e.txInfo.Nonce != nonce {
			return errors.New("invalid Nonce")
		}
	}
	return nil
}

func (e *TransferExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]
	toAccount := bc.accountMap[txInfo.ToAccountIndex]

	if err := e.verifyExpiredAt(); err != nil {
		return err
	}

	if err := e.verifyNonce(fromAccount.AccountIndex); err != nil {
		return err
	}

	if txInfo.ToAccountNameHash != toAccount.AccountNameHash {
		return errors.New("invalid to account name hash")
	}
	if txInfo.GasFeeAssetId != txInfo.AssetId {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("invalid gas asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			return errors.New("invalid asset amount")
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
	}

	return txInfo.VerifySignature(fromAccount.PublicKey)
}

func (e *TransferExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	e.tx.TxDetails = e.GenerateTxDetails()

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]
	toAccount := bc.accountMap[txInfo.ToAccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]

	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.ToAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	return nil
}

func (e *TransferExecutor) GeneratePubData() error {
	txInfo := e.txInfo
	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeTransfer))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedAmountBytes)
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *TransferExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}
	return bc.updateAccountTree(accounts, assets)
}

func (e *TransferExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.BlockHeight = e.bc.currentBlock.BlockHeight
	e.tx.StateRoot = e.bc.getStateRoot()
	e.tx.TxInfo = string(txInfoBytes)
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *TransferExecutor) GenerateTxDetails() []*tx.TxDetail {
	txInfo := e.txInfo
	fromAccount := e.bc.accountMap[txInfo.FromAccountIndex]
	toAccount := e.bc.accountMap[txInfo.ToAccountIndex]
	gasAccount := e.bc.accountMap[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, ffmath.Neg(txInfo.AssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	order++
	// from account asset gas
	baseBalance := fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance
	if txInfo.GasFeeAssetId == txInfo.AssetId {
		baseBalance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.AssetAmount)
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			baseBalance,
			fromAccount.AssetInfo[txInfo.AssetId].LpAmount,
			fromAccount.AssetInfo[txInfo.AssetId].OfferCanceledOrFinalized).String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	// to account asset a
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  toAccount.AccountName,
		Balance:      toAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, txInfo.AssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           toAccount.Nonce,
		CollectionNonce: toAccount.CollectionNonce,
	})
	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails
}
