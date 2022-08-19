package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
)

type AtomicMatchExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.AtomicMatchTxInfo

	buyOfferAssetId  int64
	buyOfferIndex    int64
	sellOfferAssetId int64
	sellOfferIndex   int64

	isFromBuyer bool // True when the sender's account is the same to buyer's account.
	isAssetGas  bool // True when the gas asset is the same to the buyer's asset.
}

func NewAtomicMatchExecutor(bc *BlockChain, tx *tx.Tx) TxExecutor {
	return &AtomicMatchExecutor{
		BaseExecutor: BaseExecutor{
			bc: bc,
			tx: tx,
		},
	}
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
		return errors.New("internal error")
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
		return errors.New("internal error")
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
	if err := e.bc.verifyExpiredAt(txInfo.ExpiredAt); err != nil {
		return err
	}
	if err := e.bc.verifyExpiredAt(txInfo.BuyOffer.ExpiredAt); err != nil {
		return errors.New("invalid BuyOffer.ExpiredAt")
	}
	if err := e.bc.verifyExpiredAt(txInfo.SellOffer.ExpiredAt); err != nil {
		return errors.New("invalid SellOffer.ExpiredAt")
	}

	fromAccount := bc.accountMap[txInfo.AccountIndex]
	buyAccount := bc.accountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.accountMap[txInfo.SellOffer.AccountIndex]

	if err := e.bc.verifyNonce(fromAccount.AccountIndex, txInfo.Nonce); err != nil {
		return err
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

	stateCache.pendingNewOffer = append(stateCache.pendingNewOffer, &nft.Offer{
		OfferType:    txInfo.BuyOffer.Type,
		OfferId:      txInfo.BuyOffer.OfferId,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		NftIndex:     txInfo.BuyOffer.NftIndex,
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetAmount:  txInfo.BuyOffer.AssetAmount.String(),
		ListedAt:     txInfo.BuyOffer.ListedAt,
		ExpiredAt:    txInfo.BuyOffer.ExpiredAt,
		TreasuryRate: txInfo.BuyOffer.TreasuryRate,
		Sig:          common.Bytes2Hex(txInfo.BuyOffer.Sig),
		Status:       nft.OfferFinishedStatus,
	})
	stateCache.pendingNewOffer = append(stateCache.pendingNewOffer, &nft.Offer{
		OfferType:    txInfo.SellOffer.Type,
		OfferId:      txInfo.SellOffer.OfferId,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		NftIndex:     txInfo.SellOffer.NftIndex,
		AssetId:      txInfo.SellOffer.AssetId,
		AssetAmount:  txInfo.SellOffer.AssetAmount.String(),
		ListedAt:     txInfo.SellOffer.ListedAt,
		ExpiredAt:    txInfo.SellOffer.ExpiredAt,
		TreasuryRate: txInfo.SellOffer.TreasuryRate,
		Sig:          common.Bytes2Hex(txInfo.SellOffer.Sig),
		Status:       nft.OfferFinishedStatus,
	})

	stateCache.pendingNewL2NftExchange = append(stateCache.pendingNewL2NftExchange, &nft.L2NftExchange{
		BuyerAccountIndex: txInfo.BuyOffer.AccountIndex,
		OwnerAccountIndex: txInfo.SellOffer.AccountIndex,
		NftIndex:          txInfo.BuyOffer.NftIndex,
		AssetId:           txInfo.BuyOffer.AssetId,
		AssetAmount:       txInfo.BuyOffer.AssetAmount.String(),
	})

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

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *AtomicMatchExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	matchNft := bc.nftMap[txInfo.SellOffer.NftIndex]

	copiedAccounts, err := e.bc.deepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex,
		txInfo.SellOffer.AccountIndex, txInfo.BuyOffer.AccountIndex, matchNft.CreatorAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.AccountIndex]
	buyAccount := copiedAccounts[txInfo.BuyOffer.AccountIndex]
	sellAccount := copiedAccounts[txInfo.SellOffer.AccountIndex]
	creatorAccount := copiedAccounts[matchNft.CreatorAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

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
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	// buyer asset A
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		AccountName:  buyAccount.AccountName,
		Balance:      buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, ffmath.Neg(txInfo.BuyOffer.AssetAmount), ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.AssetAmount)
	// buy offer
	order++
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.sellOfferIndex), 1)
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
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer
	// seller asset A
	order++
	accountOrder++
	sellDeltaAmount := ffmath.Sub(txInfo.SellOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount))
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  sellAccount.AccountName,
		Balance:      sellAccount.AssetInfo[txInfo.SellOffer.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, sellDeltaAmount, ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance, sellDeltaAmount)
	// sell offer
	order++
	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	sellOffer = new(big.Int).SetBit(sellOffer, int(e.sellOfferIndex), 1)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      e.sellOfferAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  sellAccount.AccountName,
		Balance:      sellAccount.AssetInfo[e.sellOfferAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			e.sellOfferAssetId, ZeroBigInt, ZeroBigInt, sellOffer).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized = sellOffer
	// creator fee
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: matchNft.CreatorAccountIndex,
		AccountName:  creatorAccount.AccountName,
		Balance:      creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.CreatorAmount, ZeroBigInt, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.CreatorAmount)
	// nft info
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      matchNft.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		Balance: commonAsset.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, matchNft.OwnerAccountIndex,
			matchNft.NftContentHash, matchNft.NftL1TokenId, matchNft.NftL1Address, matchNft.CreatorTreasuryRate, matchNft.CollectionId).String(),
		BalanceDelta: commonAsset.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, txInfo.BuyOffer.AccountIndex,
			matchNft.NftContentHash, matchNft.NftL1TokenId, matchNft.NftL1Address, matchNft.CreatorTreasuryRate, matchNft.CollectionId).String(),
		Order:        order,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	// gas account asset A - treasury fee
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.TreasuryAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.TreasuryAmount)
	// gas account asset gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	return txDetails, nil
}