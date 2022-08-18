package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"

	"github.com/bnb-chain/zkbas/common/util"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

type AtomicMatchExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *legendTxTypes.AtomicMatchTxInfo

	buyOfferAssetId  int64
	buyOfferIndex    int64
	sellOfferAssetId int64
	sellOfferIndex   int64

	isFromBuyer bool // True when the sender's account is the same to buyer's account.
	isAssetGas  bool // True when the gas asset is the same to the buyer's asset.
}

func NewAtomicMatchExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &AtomicMatchExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *AtomicMatchExecutor) Prepare() error {
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	e.buyOfferAssetId = txInfo.BuyOffer.OfferId / txVerification.OfferPerAsset
	e.buyOfferIndex = txInfo.BuyOffer.OfferId % txVerification.OfferPerAsset
	e.sellOfferAssetId = txInfo.SellOffer.OfferId / txVerification.OfferPerAsset
	e.sellOfferIndex = txInfo.SellOffer.OfferId % txVerification.OfferPerAsset

	// Prepare seller's asset and nft, if the buyer's asset or nft isn't the same, it will be failed in the verify step.
	err = e.bc.prepareNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return err
	}

	matchNft := e.bc.nftMap[txInfo.SellOffer.NftIndex]
	e.isFromBuyer = true
	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex, txInfo.SellOffer.AccountIndex, matchNft.CreatorAccountIndex}
	if txInfo.AccountIndex != txInfo.BuyOffer.AccountIndex {
		e.isFromBuyer = false
		accounts = append(accounts, txInfo.BuyOffer.AccountIndex)
	}
	e.isAssetGas = true
	assets := []int64{txInfo.GasFeeAssetId, e.buyOfferAssetId, e.sellOfferAssetId}
	if txInfo.GasFeeAssetId != txInfo.SellOffer.AssetId {
		e.isAssetGas = false
		assets = append(assets, txInfo.SellOffer.AssetId)
	}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	// Set the right treasury and creator treasury amount.
	txInfo.TreasuryAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.TreasuryRate)), big.NewInt(txVerification.TenThousand))
	txInfo.CreatorAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(matchNft.CreatorTreasuryRate)), big.NewInt(txVerification.TenThousand))

	e.txInfo = txInfo
	return nil
}

func (e *AtomicMatchExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}
	if txInfo.BuyOffer.Type != commonAsset.BuyOfferType ||
		txInfo.SellOffer.Type != commonAsset.SellOfferType {
		return errors.New("invalid offer type")
	}
	if txInfo.BuyOffer.AccountIndex == txInfo.SellOffer.AccountIndex {
		return errors.New("same buyer and seller")
	}
	if txInfo.SellOffer.NftIndex != txInfo.BuyOffer.NftIndex ||
		txInfo.SellOffer.AssetId != txInfo.BuyOffer.AssetId ||
		txInfo.SellOffer.AssetAmount.String() != txInfo.BuyOffer.AssetAmount.String() ||
		txInfo.SellOffer.TreasuryRate != txInfo.BuyOffer.TreasuryRate {
		return errors.New("buy offer mismatches sell offer")
	}

	// Check expired time.
	if txInfo.ExpiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("tx expired")
	}
	if txInfo.SellOffer.ExpiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("sell offer expired")
	}
	if txInfo.BuyOffer.ExpiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("buy offer expired")
	}

	fromAccount := bc.accountMap[txInfo.AccountIndex]
	buyAccount := bc.accountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.accountMap[txInfo.SellOffer.AccountIndex]

	// Check from account's nonce.
	if txInfo.Nonce != fromAccount.Nonce {
		return errors.New("invalid nonce")
	}

	// Check sender's gas balance and buyer's asset balance.
	if e.isFromBuyer && e.isAssetGas {
		totalBalance := ffmath.Add(txInfo.GasFeeAssetAmount, txInfo.BuyOffer.AssetAmount)
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(totalBalance) < 0 {
			return errors.New("sender balance is not enough")
		}
	} else {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("sender balance is not enough")
		}

		if buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance.Cmp(txInfo.BuyOffer.AssetAmount) < 0 {
			return errors.New("buy balance is not enough")
		}
	}

	// Check offer canceled or finalized.
	sellOffer := bc.accountMap[txInfo.SellOffer.AccountIndex].AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	if sellOffer.Bit(int(e.sellOfferIndex)) == 1 {
		return errors.New("sell offer canceled or finalized")
	}
	buyOffer := bc.accountMap[txInfo.BuyOffer.AccountIndex].AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	if buyOffer.Bit(int(e.buyOfferIndex)) == 1 {
		return errors.New("buy offer canceled or finalized")
	}

	// Check the seller is the owner of the nft.
	if bc.nftMap[txInfo.SellOffer.NftIndex].OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		return errors.New("seller is not owner")
	}

	err = txInfo.BuyOffer.VerifySignature(buyAccount.PublicKey)
	if err != nil {
		return err
	}
	err = txInfo.SellOffer.VerifySignature(sellAccount.PublicKey)
	if err != nil {
		return err
	}
	err = txInfo.VerifySignature(fromAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (e *AtomicMatchExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// generate tx details
	e.tx.TxDetails = e.GenerateTxDetails()

	// apply changes
	matchNft := bc.nftMap[txInfo.SellOffer.NftIndex]
	fromAccount := bc.accountMap[txInfo.AccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]
	buyAccount := bc.accountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.accountMap[txInfo.SellOffer.AccountIndex]
	creatorAccount := bc.accountMap[matchNft.CreatorAccountIndex]

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.AssetAmount)
	sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance, ffmath.Sub(
		txInfo.BuyOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount)))
	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.CreatorAmount)
	gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.TreasuryAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	sellOffer = new(big.Int).SetBit(sellOffer, int(e.sellOfferIndex), 1)
	sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized = sellOffer
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.sellOfferIndex), 1)
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer

	// Change new owner.
	matchNft.OwnerAccountIndex = txInfo.BuyOffer.AccountIndex

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.BuyOffer.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.SellOffer.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[matchNft.CreatorAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingUpdateNftIndexMap[txInfo.SellOffer.NftIndex] = StateCachePending
	return nil
}

func (e *AtomicMatchExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeAtomicMatch))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
	buf.Write(util.Uint24ToBytes(txInfo.BuyOffer.OfferId))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
	buf.Write(util.Uint24ToBytes(txInfo.SellOffer.OfferId))
	buf.Write(util.Uint40ToBytes(txInfo.BuyOffer.NftIndex))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	packedAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAmountBytes)
	creatorAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.CreatorAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(creatorAmountBytes)
	treasuryAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(treasuryAmountBytes)
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
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *AtomicMatchExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	matchNft := e.bc.nftMap[txInfo.SellOffer.NftIndex]
	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex, txInfo.SellOffer.AccountIndex, matchNft.CreatorAccountIndex}
	if !e.isFromBuyer {
		accounts = append(accounts, txInfo.BuyOffer.AccountIndex)
	}
	assets := []int64{txInfo.GasFeeAssetId, e.buyOfferAssetId, e.sellOfferAssetId}
	if !e.isAssetGas {
		assets = append(assets, txInfo.SellOffer.AssetId)
	}
	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.updateNftTree(txInfo.SellOffer.NftIndex)
	if err != nil {
		logx.Errorf("update nft tree error, err: %s", err.Error())
		return err
	}
	return nil
}

func (e *AtomicMatchExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *AtomicMatchExecutor) GenerateTxDetails() []*tx.TxDetail {
	bc := e.bc
	txInfo := e.txInfo
	//matchNft := bc.nftMap[txInfo.SellOffer.NftIndex]
	fromAccount := bc.accountMap[txInfo.AccountIndex]
	//gasAccount := bc.accountMap[txInfo.GasAccountIndex]
	buyAccount := bc.accountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.accountMap[txInfo.SellOffer.AccountIndex]
	//creatorAccount := bc.accountMap[matchNft.CreatorAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 9)

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
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromGasBalance := ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	// buyer asset A
	order++
	accountOrder++
	buyAssetBalance := buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance
	if e.isFromBuyer && e.isAssetGas {
		buyAssetBalance = fromGasBalance
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		AccountName:  buyAccount.AccountName,
		Balance: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId,
			buyAssetBalance,
			buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].LpAmount,
			buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].OfferCanceledOrFinalized).String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, ffmath.Neg(txInfo.BuyOffer.AssetAmount), ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	buyAssetBalance = ffmath.Sub(buyAssetBalance, txInfo.BuyOffer.AssetAmount)
	if e.isFromBuyer && e.isAssetGas {
		fromGasBalance = buyAssetBalance
	}
	// buy offer
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.sellOfferIndex), 1)
	if e.buyOfferAssetId == txInfo.BuyOffer.AssetId {
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:      e.buyOfferAssetId,
			AssetType:    commonAsset.GeneralAssetType,
			AccountIndex: txInfo.BuyOffer.AccountIndex,
			AccountName:  buyAccount.AccountName,
			Balance: commonAsset.ConstructAccountAsset(
				e.buyOfferAssetId,
				buyAssetBalance,
				buyAccount.AssetInfo[e.buyOfferAssetId].LpAmount,
				fromAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized).String(),
			BalanceDelta: commonAsset.ConstructAccountAsset(
				e.buyOfferAssetId, ZeroBigInt, ZeroBigInt, buyOffer).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	} else {
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:      e.buyOfferAssetId,
			AssetType:    commonAsset.GeneralAssetType,
			AccountIndex: txInfo.BuyOffer.AccountIndex,
			AccountName:  buyAccount.AccountName,
			Balance:      buyAccount.AssetInfo[e.buyOfferAssetId].String(),
			BalanceDelta: commonAsset.ConstructAccountAsset(
				e.buyOfferAssetId, ZeroBigInt, ZeroBigInt, buyOffer).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	}
	// seller asset A
	sellAssetBalance := sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance
	if txInfo.AccountIndex == txInfo.SellOffer.AccountIndex && e.isAssetGas {
		sellAssetBalance = fromGasBalance
	}
	sellDeltaAmount := ffmath.Sub(txInfo.SellOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount))
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  sellAccount.AccountName,
		Balance: commonAsset.ConstructAccountAsset(
			txInfo.SellOffer.AssetId,
			sellAssetBalance,
			sellAccount.AssetInfo[txInfo.SellOffer.AssetId].LpAmount,
			sellAccount.AssetInfo[txInfo.SellOffer.AssetId].OfferCanceledOrFinalized).String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, sellDeltaAmount, ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	sellAssetBalance = ffmath.Add(sellAssetBalance, sellDeltaAmount)
	if txInfo.AccountIndex == txInfo.SellOffer.AccountIndex && e.isAssetGas {
		fromGasBalance = sellAssetBalance
	}
	// sell offer
	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	sellOffer = new(big.Int).SetBit(sellOffer, int(e.sellOfferIndex), 1)
	if e.sellOfferAssetId == txInfo.SellOffer.AssetId {
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:      e.sellOfferAssetId,
			AssetType:    commonAsset.GeneralAssetType,
			AccountIndex: txInfo.SellOffer.AccountIndex,
			AccountName:  sellAccount.AccountName,
			Balance: commonAsset.ConstructAccountAsset(
				e.sellOfferAssetId,
				sellAssetBalance,
				sellAccount.AssetInfo[e.sellOfferAssetId].LpAmount,
				sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized).String(),
			BalanceDelta: commonAsset.ConstructAccountAsset(
				e.sellOfferAssetId, ZeroBigInt, ZeroBigInt, sellOffer).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	} else {
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:      e.sellOfferAssetId,
			AssetType:    commonAsset.GeneralAssetType,
			AccountIndex: txInfo.SellOffer.AccountIndex,
			AccountName:  sellAccount.AccountName,
			Balance:      sellAccount.AssetInfo[e.sellOfferAssetId].String(),
			BalanceDelta: commonAsset.ConstructAccountAsset(
				e.sellOfferAssetId, ZeroBigInt, ZeroBigInt, buyOffer).String(),
			Order:        order,
			AccountOrder: accountOrder,
		})
	}
	// TODO: maybe need bettor method to generate tx details.

	return txDetails
}
