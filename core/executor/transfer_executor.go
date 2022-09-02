package executor

import (
	"bytes"
	"encoding/json"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type TransferExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.TransferTxInfo
}

func NewTransferExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseTransferTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &TransferExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *TransferExecutor) Prepare() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetId, txInfo.GasFeeAssetId}
	err := e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	return nil
}

func (e *TransferExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	toAccount := bc.StateDB().AccountMap[txInfo.ToAccountIndex]
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

	return nil
}

func (e *TransferExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	toAccount := bc.StateDB().AccountMap[txInfo.ToAccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]

	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.ToAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	stateCache.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetId})
	stateCache.MarkAccountAssetsDirty(txInfo.ToAccountIndex, []int64{txInfo.AssetId})
	stateCache.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	return nil
}

func (e *TransferExecutor) GeneratePubData() error {
	txInfo := e.txInfo
	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeTransfer))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedAmountBytes)
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *TransferExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *TransferExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex, txInfo.ToAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]
	toAccount := copiedAccounts[txInfo.ToAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetId, ffmath.Neg(txInfo.AssetAmount), types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	// from account asset gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	// to account asset a
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  toAccount.AccountName,
		Balance:      toAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetId, txInfo.AssetAmount, types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           toAccount.Nonce,
		CollectionNonce: toAccount.CollectionNonce,
	})
	toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	return txDetails, nil
}

func (e *TransferExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := e.txInfo.Hash(mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      types.NilNftIndex,
		PairIndex:     types.NilPairIndex,
		AssetId:       types.NilAssetId,
		TxAmount:      e.txInfo.AssetAmount.String(),
		Memo:          e.txInfo.Memo,
		AccountIndex:  e.txInfo.FromAccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
