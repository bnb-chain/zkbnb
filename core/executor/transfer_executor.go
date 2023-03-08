package executor

import (
	"bytes"
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/ethereum/go-ethereum/common"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type TransferExecutor struct {
	BaseExecutor

	TxInfo          *txtypes.TransferTxInfo
	IsCreateAccount bool
}

func NewTransferExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseTransferTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &TransferExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		TxInfo:       txInfo,
	}, nil
}

func (e *TransferExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.TxInfo
	toL1Address := txInfo.ToL1Address
	toAccount, err := bc.StateDB().GetAccountByL1Address(toL1Address)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	if err == types.DbErrNotFound {
		if txInfo.ToAccountIndex != -1 {
			return types.AppErrAccountInvalidToAccount
		}
		if !e.bc.StateDB().DryRun {
			if e.tx.Rollback == false {
				nextAccountIndex := e.bc.StateDB().GetNextAccountIndex()
				txInfo.ToAccountIndex = nextAccountIndex
			} else {
				//for rollback
				txInfo.ToAccountIndex = e.tx.AccountIndex
			}
		}
		e.IsCreateAccount = true
	} else {
		if txInfo.ToAccountIndex != toAccount.AccountIndex {
			return types.AppErrInvalidToAddress
		}
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetId})
	e.MarkAccountAssetsDirty(txInfo.ToAccountIndex, []int64{txInfo.AssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	return e.BaseExecutor.Prepare()
}

func (e *TransferExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	bc := e.bc
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	if !e.IsCreateAccount {
		toAccount, err := bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
		if err != nil {
			return err
		}
		if fromAccount.AccountIndex == toAccount.AccountIndex {
			return types.AppErrAccountInvalidToAccount
		}
		if txInfo.ToL1Address != toAccount.L1Address {
			return types.AppErrInvalidToAddress
		}
	}

	if txInfo.GasFeeAssetId != txInfo.AssetId {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return types.AppErrInvalidGasFeeAmount
		}
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			return types.AppErrInvalidAssetAmount
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(deltaBalance) < 0 {
			return types.AppErrInvalidAssetAmount
		}
	}

	return nil
}

func (e *TransferExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	toAccount, err := bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	if err == types.DbErrNotFound {
		newAccount := &account.Account{
			AccountIndex:    txInfo.ToAccountIndex,
			PublicKey:       types.EmptyPk,
			L1Address:       e.TxInfo.ToL1Address,
			Nonce:           types.EmptyNonce,
			CollectionNonce: types.EmptyCollectionNonce,
			AssetInfo:       types.EmptyAccountAssetInfo,
			AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
			Status:          account.AccountStatusPending,
		}
		toAccount, err = chain.ToFormatAccountInfo(newAccount)
		if err != nil {
			return err
		}
	}

	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.FromAccountIndex, fromAccount)
	stateCache.SetPendingAccount(txInfo.ToAccountIndex, toAccount)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *TransferExecutor) GeneratePubData() error {
	txInfo := e.TxInfo
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
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *TransferExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.AssetId = e.TxInfo.AssetId
	e.tx.TxAmount = e.TxInfo.AssetAmount.String()
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *TransferExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

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
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetId, ffmath.Neg(txInfo.AssetAmount), types.ZeroBigInt).String(),
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
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), types.ZeroBigInt).String(),
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
		L1Address:    toAccount.L1Address,
		Balance:      toAccount.AssetInfo[txInfo.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetId, txInfo.AssetAmount, types.ZeroBigInt).String(),
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
		L1Address:    gasAccount.L1Address,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
	})
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	return txDetails, nil
}

func (e *TransferExecutor) Finalize() error {
	if e.IsCreateAccount {
		bc := e.bc
		txInfo := e.TxInfo
		bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.ToAccountIndex, bc.CurrentBlock().BlockHeight)
	}
	return nil
}

