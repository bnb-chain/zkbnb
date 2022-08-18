package core

import (
	"bytes"
	"encoding/json"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type WithdrawExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *legendTxTypes.WithdrawTxInfo
}

func NewWithdrawExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &WithdrawExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *WithdrawExecutor) Prepare() error {
	txInfo, err := commonTx.ParseWithdrawTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	accounts := []int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	e.txInfo = txInfo
	return nil
}

func (e *WithdrawExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	if txInfo.ExpiredAt < e.bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("tx expired")
	}

	fromAccount := e.bc.accountMap[txInfo.FromAccountIndex]
	if txInfo.Nonce != fromAccount.Nonce {
		return errors.New("invalid nonce")
	}

	if txInfo.GasFeeAssetId != txInfo.AssetId {
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			return errors.New("invalid asset amount")
		}
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("invalid gas asset amount")
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
	}

	err = txInfo.VerifySignature(fromAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (e *WithdrawExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]

	// apply changes
	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.priorityOperations++

	return nil
}

func (e *WithdrawExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeWithdrawNft))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetId)))
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(util.Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk2 := util.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pendingOnChainOperationsPubData = append(stateCache.pendingOnChainOperationsPubData, pubData)
	stateCache.pendingOnChainOperationsHash = util.ConcatKeccakHash(stateCache.pendingOnChainOperationsHash, pubData)
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *WithdrawExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}

	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	return nil
}

func (e *WithdrawExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *WithdrawExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	fromAccount := e.bc.accountMap[txInfo.FromAccountIndex]
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
	return txDetails, nil
}
