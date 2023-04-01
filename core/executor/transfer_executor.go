package executor

import (
	"bytes"
	"encoding/json"
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
		return nil, types.AppErrInvalidTxInfo
	}

	return &TransferExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func (e *TransferExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.TxInfo
	toL1Address := txInfo.ToL1Address
	if !e.isDesertExit {
		txInfo.ToAccountIndex = types.NilAccountIndex
	}
	toAccount, err := bc.StateDB().GetAccountByL1Address(toL1Address)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	if err == types.AppErrAccountNotFound {
		if !e.isDesertExit {
			if !e.bc.StateDB().IsFromApi {
				if e.tx.Rollback == false {
					nextAccountIndex := e.bc.StateDB().GetNextAccountIndex()
					txInfo.ToAccountIndex = nextAccountIndex
				} else {
					//for rollback
					txInfo.ToAccountIndex = e.tx.ToAccountIndex
				}
			}
		}
		e.IsCreateAccount = true
	} else {
		txInfo.ToAccountIndex = toAccount.AccountIndex
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetId})
	if e.IsCreateAccount {
		err := e.CreateEmptyAccount(txInfo.ToAccountIndex, txInfo.ToL1Address, []int64{txInfo.AssetId})
		if err != nil {
			return err
		}
	}
	// Mark the tree states that would be affected in this executor.
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
	var toAccount *types.AccountInfo
	var err error

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	if e.IsCreateAccount {
		toAccount = e.GetCreatingAccount()
	} else {
		toAccount, err = bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
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
	buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
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
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.AssetId = e.TxInfo.AssetId
	e.tx.TxAmount = e.TxInfo.AssetAmount.String()
	e.tx.IsCreateAccount = e.IsCreateAccount
	if e.tx.ToAccountIndex != e.iTxInfo.GetToAccountIndex() || e.IsCreateAccount {
		e.tx.IsPartialUpdate = true
	}
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *TransferExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo
	var copiedAccounts map[int64]*types.AccountInfo
	var err error
	var toAccount *types.AccountInfo
	if e.IsCreateAccount {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = e.GetEmptyAccount()
	} else {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex, txInfo.ToAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = copiedAccounts[txInfo.ToAccountIndex]
	}

	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

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
		PublicKey:       fromAccount.PublicKey,
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
		PublicKey:       fromAccount.PublicKey,
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
		PublicKey:       toAccount.PublicKey,
	})
	toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	if e.IsCreateAccount {
		order++
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:         types.EmptyAccountAssetId,
			AssetType:       types.CreateAccountType,
			AccountIndex:    txInfo.ToAccountIndex,
			L1Address:       toAccount.L1Address,
			Balance:         toAccount.L1Address,
			BalanceDelta:    txInfo.ToL1Address,
			Order:           order,
			AccountOrder:    accountOrder,
			Nonce:           toAccount.Nonce,
			CollectionNonce: toAccount.CollectionNonce,
			PublicKey:       toAccount.PublicKey,
		})
	}
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
		PublicKey:       gasAccount.PublicKey,
	})
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	return txDetails, nil
}

func (e *TransferExecutor) Finalize() error {
	if e.IsCreateAccount {
		bc := e.bc
		txInfo := e.TxInfo
		if !e.isDesertExit {
			bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.ToAccountIndex, bc.CurrentBlock().BlockHeight)
		}
		accountInfo := e.GetCreatingAccount()
		bc.StateDB().SetPendingAccountL1AddressMap(accountInfo.L1Address, accountInfo.AccountIndex)
	}
	err := e.BaseExecutor.Finalize()
	if err != nil {
		return err
	}
	return nil
}
