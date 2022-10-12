package executor

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type CreateCollectionExecutor struct {
	BaseExecutor

	txInfo *txtypes.CreateCollectionTxInfo
}

func NewCreateCollectionExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseCreateCollectionTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &CreateCollectionExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *CreateCollectionExecutor) Prepare() error {
	txInfo := e.txInfo

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	err := e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Set the right collection nonce to tx info.
	account, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	txInfo.CollectionId = account.CollectionNonce
	return nil
}

func (e *CreateCollectionExecutor) VerifyInputs(skipGasAmtChk bool) error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	return nil
}

func (e *CreateCollectionExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}

	// apply changes
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++
	fromAccount.CollectionNonce++

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(fromAccount.AccountIndex, fromAccount)
	stateCache.SetPendingUpdateGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *CreateCollectionExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeCreateCollection))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *CreateCollectionExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.txInfo.GasFeeAssetId
	e.tx.GasFee = e.txInfo.GasFeeAssetAmount.String()
	e.tx.CollectionId = e.txInfo.CollectionId
	return e.BaseExecutor.GetExecutedTx()
}

func (e *CreateCollectionExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}

	fromAccount := copiedAccounts[txInfo.AccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account collection nonce
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         types.NilAssetId,
		AssetType:       types.CollectionNonceAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     fromAccount.AccountName,
		Balance:         strconv.FormatInt(fromAccount.CollectionNonce, 10),
		BalanceDelta:    strconv.FormatInt(fromAccount.CollectionNonce+1, 10),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.CollectionNonce = fromAccount.CollectionNonce + 1

	// from account gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
		).String(),
		Order:        order,
		Nonce:        gasAccount.Nonce,
		AccountOrder: accountOrder,
		IsGas:        true,
	})
	return txDetails, nil
}
