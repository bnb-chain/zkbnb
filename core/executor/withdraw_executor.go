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

type WithdrawExecutor struct {
	BaseExecutor

	TxInfo *txtypes.WithdrawTxInfo
}

func NewWithdrawExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseWithdrawTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &WithdrawExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewWithdrawExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &WithdrawExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.WithdrawTxInfo),
	}, nil
}

func (e *WithdrawExecutor) Prepare() error {
	txInfo := e.TxInfo

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	return e.BaseExecutor.Prepare()
}

func (e *WithdrawExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}

	if txInfo.GasFeeAssetId != txInfo.AssetId {
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			return types.AppErrInvalidAssetAmount
		}
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return types.AppErrInvalidGasFeeAccount
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(deltaBalance) < 0 {
			return types.AppErrInvalidAssetAmount
		}
	}

	return nil
}

func (e *WithdrawExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}

	// apply changes
	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.FromAccountIndex, fromAccount)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *WithdrawExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeWithdraw))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *WithdrawExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
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
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *WithdrawExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 3)
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
	if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientAssetBalance
	}

	order++
	// from account asset gas
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
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientGasFeeBalance
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
	return txDetails, nil
}
