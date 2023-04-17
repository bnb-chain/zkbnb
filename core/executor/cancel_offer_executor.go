package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type CancelOfferExecutor struct {
	BaseExecutor

	TxInfo *txtypes.CancelOfferTxInfo
}

func NewCancelOfferExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseCancelOfferTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &CancelOfferExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewCancelOfferExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &CancelOfferExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.CancelOfferTxInfo),
	}, nil
}

func (e *CancelOfferExecutor) PreLoadAccountAndNft(accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
	txInfo := e.TxInfo
	accountIndexMap[txInfo.AccountIndex] = true
	accountIndexMap[txInfo.GasAccountIndex] = true
}

func (e *CancelOfferExecutor) Prepare() error {
	txInfo := e.TxInfo

	// Mark the tree states that would be affected in this executor.
	offerAssetId := txInfo.OfferId / OfferPerAsset
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId, offerAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	return e.BaseExecutor.Prepare()
}

func (e *CancelOfferExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}

	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}

	offerAssetId := txInfo.OfferId / 128
	offerIndex := txInfo.OfferId % 128
	offerAsset := fromAccount.AssetInfo[offerAssetId]
	if offerAsset != nil && offerAsset.OfferCanceledOrFinalized != nil {
		xBit := offerAsset.OfferCanceledOrFinalized.Bit(int(offerIndex))
		if xBit == 1 {
			return types.AppErrInvalidOfferState
		}
	}

	return nil
}

func (e *CancelOfferExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

	// apply changes
	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	offerAssetId := txInfo.OfferId / OfferPerAsset
	offerIndex := txInfo.OfferId % OfferPerAsset
	oOffer := fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	nOffer := new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(fromAccount.AccountIndex, fromAccount)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *CancelOfferExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeCancelOffer))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint24ToBytes(txInfo.OfferId))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *CancelOfferExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *CancelOfferExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.AccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		L1Address:    fromAccount.L1Address,
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
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientGasFeeBalance
	}

	// from account offer id
	offerAssetId := txInfo.OfferId / OfferPerAsset
	offerIndex := txInfo.OfferId % OfferPerAsset
	oldOffer := fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	// verify whether account offer id is valid for use
	if oldOffer.Bit(int(offerIndex)) == 1 {
		logx.Errorf("account %d offer index %d is already in use", txInfo.AccountIndex, offerIndex)
		return nil, types.AppErrOfferIndexAlreadyInUse
	}
	nOffer := new(big.Int).SetBit(oldOffer, int(offerIndex), 1)

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      offerAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[offerAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			offerAssetId,
			types.ZeroBigInt,
			nOffer,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		L1Address:    gasAccount.L1Address,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
		PublicKey:       gasAccount.PublicKey,
	})
	return txDetails, nil
}
