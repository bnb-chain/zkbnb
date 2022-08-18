package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
)

type CancelOfferExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *legendTxTypes.CancelOfferTxInfo
}

func NewCancelOfferExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &CancelOfferExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *CancelOfferExecutor) Prepare() error {
	txInfo, err := commonTx.ParseCancelOfferTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	offerAssetId := txInfo.OfferId / txVerification.OfferPerAsset

	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex}
	assets := []int64{offerAssetId, txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	e.txInfo = txInfo
	return nil
}

func (e *CancelOfferExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	if err := e.bc.verifyExpiredAt(txInfo.ExpiredAt); err != nil {
		return err
	}

	fromAccount := e.bc.accountMap[txInfo.AccountIndex]

	if err := e.bc.verifyNonce(fromAccount.AccountIndex, txInfo.Nonce); err != nil {
		return err
	}

	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	err = txInfo.VerifySignature(fromAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (e *CancelOfferExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// apply changes
	fromAccount := bc.accountMap[txInfo.AccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	offerAssetId := txInfo.OfferId / txVerification.OfferPerAsset
	offerIndex := txInfo.OfferId % txVerification.OfferPerAsset
	oOffer := fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	nOffer := new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending

	stateCache.pendingNewOffer = append(stateCache.pendingNewOffer, &nft.Offer{
		OfferType:    0,
		OfferId:      txInfo.OfferId,
		AccountIndex: txInfo.AccountIndex,
		NftIndex:     0,
		AssetId:      0,
		AssetAmount:  "0",
		ListedAt:     0,
		ExpiredAt:    0,
		TreasuryRate: 0,
		Sig:          "",
		Status:       nft.OfferFinishedStatus,
	})

	return nil
}

func (e *CancelOfferExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeCancelOffer))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint24ToBytes(txInfo.OfferId))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *CancelOfferExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	offerAssetId := txInfo.OfferId / txVerification.OfferPerAsset
	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex}
	assets := []int64{offerAssetId, txInfo.GasFeeAssetId}

	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	return nil
}

func (e *CancelOfferExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *CancelOfferExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	copiedAccounts, err := e.bc.deepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex})
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
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// from account offer id
	offerAssetId := txInfo.OfferId / txVerification.OfferPerAsset
	offerIndex := txInfo.OfferId % txVerification.OfferPerAsset
	oldOffer := fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	// verify whether account offer id is valid for use
	if oldOffer.Bit(int(offerIndex)) == 1 {
		logx.Errorf("account %d offer index %d is already in use", txInfo.AccountIndex, offerIndex)
		return nil, errors.New("unexpected err")
	}
	nOffer := new(big.Int).SetBit(oldOffer, int(offerIndex), 1)

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      offerAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[offerAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			offerAssetId,
			ZeroBigInt,
			ZeroBigInt,
			nOffer,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails, nil
}
