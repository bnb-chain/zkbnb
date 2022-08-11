package core

import (
	"bytes"
	"encoding/json"

	"github.com/bnb-chain/zkbas/common/util"

	"github.com/bnb-chain/zkbas/common/commonAsset"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type TransferExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *TransferTxInfo
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

func (e *TransferExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo
	fromAccountInfo := bc.AccountMap[txInfo.FromAccountIndex]

	if txInfo.AssetAmount.Cmp(ZeroBigInt) <= 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) <= 0 ||
		fromAccountInfo.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 ||
		fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("invalid params")
	}

	if txInfo.Nonce != fromAccountInfo.Nonce {
		return errors.New("invalid nonce")
	}

	return nil
}

func (e *TransferExecutor) ApplyTransaction(stateCache *StateCache) (*StateCache, error) {
	bc := e.bc
	txInfo := e.txInfo
	fromAccountInfo := bc.AccountMap[txInfo.FromAccountIndex]
	toAccountInfo := bc.AccountMap[txInfo.ToAccountIndex]
	gasAccountInfo := bc.AccountMap[txInfo.GasAccountIndex]

	fromAccountInfo.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	toAccountInfo.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccountInfo.Nonce++

	stateCache.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.ToAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	return stateCache, nil
}

func (e *TransferExecutor) GeneratePubData(stateCache *StateCache) (*StateCache, error) {
	txInfo := e.txInfo
	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeTransfer))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		return stateCache, err
	}
	buf.Write(packedAmountBytes)
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return stateCache, err
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
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return stateCache, nil
}

func (e *TransferExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}
	return bc.updateAccountTree(accounts, assets)
}

func (e *TransferExecutor) GetExecutedTx() (*tx.Tx, error) {
	bc := e.bc
	txInfo := e.txInfo
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}
	stateRoot := bc.getStateRoot()
	e.tx.BlockHeight = bc.currentBlock
	e.tx.StateRoot = stateRoot
	e.tx.TxInfo = string(txInfoBytes)
	e.tx.TxStatus = tx.StatusPending
	e.tx.TxDetails = e.GenerateTxDetails()
	return e.tx, nil
}

func (e *TransferExecutor) GenerateTxDetails() []*tx.TxDetail {
	bc := e.bc
	txInfo := e.txInfo
	txDetails := make([]*tx.TxDetail, 0, 4)
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  bc.AccountMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, ffmath.Neg(txInfo.AssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	order++
	// from account asset gas
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  bc.AccountMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// to account asset a
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  bc.AccountMap[txInfo.ToAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetId, txInfo.AssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  bc.AccountMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	return txDetails
}
