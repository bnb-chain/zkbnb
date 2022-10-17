package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type AtomicMatchExecutor struct {
	BaseExecutor

	txInfo *txtypes.AtomicMatchTxInfo

	buyOfferAssetId  int64
	buyOfferIndex    int64
	sellOfferAssetId int64
	sellOfferIndex   int64
}

func NewAtomicMatchExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseAtomicMatchTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &AtomicMatchExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *AtomicMatchExecutor) Prepare() error {
	txInfo := e.txInfo

	e.buyOfferAssetId = txInfo.BuyOffer.OfferId / OfferPerAsset
	e.buyOfferIndex = txInfo.BuyOffer.OfferId % OfferPerAsset
	e.sellOfferAssetId = txInfo.SellOffer.OfferId / OfferPerAsset
	e.sellOfferIndex = txInfo.SellOffer.OfferId % OfferPerAsset

	// Prepare seller's asset and nft, if the buyer's asset or nft isn't the same, it will be failed in the verify step.
	matchNft, err := e.bc.StateDB().PrepareNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return err
	}

	// Set the right treasury and creator treasury amount.
	txInfo.TreasuryAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.TreasuryRate)), big.NewInt(TenThousand))
	txInfo.CreatorAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(matchNft.CreatorTreasuryRate)), big.NewInt(TenThousand))

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.SellOffer.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.BuyOffer.AccountIndex, []int64{txInfo.BuyOffer.AssetId, e.buyOfferAssetId})
	e.MarkAccountAssetsDirty(txInfo.SellOffer.AccountIndex, []int64{txInfo.SellOffer.AssetId, e.sellOfferAssetId})
	e.MarkAccountAssetsDirty(matchNft.CreatorAccountIndex, []int64{txInfo.BuyOffer.AssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.BuyOffer.AssetId, txInfo.GasFeeAssetId})
	return e.BaseExecutor.Prepare()
}

func (e *AtomicMatchExecutor) VerifyInputs(skipGasAmtChk bool) error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk)
	if err != nil {
		return err
	}

	if txInfo.BuyOffer.Type != types.BuyOfferType ||
		txInfo.SellOffer.Type != types.SellOfferType {
		return types.AppErrInvalidOfferType
	}
	if txInfo.BuyOffer.AccountIndex == txInfo.SellOffer.AccountIndex {
		return types.AppErrSameBuyerAndSeller
	}
	if txInfo.SellOffer.NftIndex != txInfo.BuyOffer.NftIndex ||
		txInfo.SellOffer.AssetId != txInfo.BuyOffer.AssetId ||
		txInfo.SellOffer.AssetAmount.String() != txInfo.BuyOffer.AssetAmount.String() ||
		txInfo.SellOffer.TreasuryRate != txInfo.BuyOffer.TreasuryRate {
		return types.AppErrBuyOfferMismatchSellOffer
	}

	// only gas assets are allowed for atomic match
	found := false
	for _, assetId := range types.GasAssets {
		if assetId == txInfo.SellOffer.AssetId {
			found = true
		}
	}
	if !found {
		return types.AppErrInvalidAssetOfOffer
	}

	// Check offer expired time.
	if err := e.bc.VerifyExpiredAt(txInfo.BuyOffer.ExpiredAt); err != nil {
		return types.AppErrInvalidBuyOfferExpireTime
	}
	if err := e.bc.VerifyExpiredAt(txInfo.SellOffer.ExpiredAt); err != nil {
		return types.AppErrInvalidSellOfferExpireTime
	}

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	buyAccount, err := bc.StateDB().GetFormatAccount(txInfo.BuyOffer.AccountIndex)
	if err != nil {
		return err
	}
	sellAccount, err := bc.StateDB().GetFormatAccount(txInfo.SellOffer.AccountIndex)
	if err != nil {
		return err
	}

	// Check sender's gas balance and buyer's asset balance.
	if txInfo.AccountIndex == txInfo.BuyOffer.AccountIndex && txInfo.GasFeeAssetId == txInfo.SellOffer.AssetId {
		totalBalance := ffmath.Add(txInfo.GasFeeAssetAmount, txInfo.BuyOffer.AssetAmount)
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(totalBalance) < 0 {
			return types.AppErrSellerBalanceNotEnough
		}
	} else {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return types.AppErrSellerBalanceNotEnough
		}

		if buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance.Cmp(txInfo.BuyOffer.AssetAmount) < 0 {
			return types.AppErrBuyerBalanceNotEnough
		}
	}

	// Check offer canceled or finalized.
	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	if sellOffer.Bit(int(e.sellOfferIndex)) == 1 {
		return types.AppErrInvalidSellOfferState
	}
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	if buyOffer.Bit(int(e.buyOfferIndex)) == 1 {
		return types.AppErrInvalidBuyOfferState
	}

	// Check the seller is the owner of the nft.
	nft, err := bc.StateDB().GetNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		return err
	}
	if nft.OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		return types.AppErrSellerNotOwner
	}

	// Verify offer signature.
	err = txInfo.BuyOffer.VerifySignature(buyAccount.PublicKey)
	if err != nil {
		return err
	}
	err = txInfo.SellOffer.VerifySignature(sellAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (e *AtomicMatchExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// apply changes
	matchNft, err := bc.StateDB().GetNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		return err
	}
	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	buyAccount, err := bc.StateDB().GetFormatAccount(txInfo.BuyOffer.AccountIndex)
	if err != nil {
		return err
	}
	sellAccount, err := bc.StateDB().GetFormatAccount(txInfo.SellOffer.AccountIndex)
	if err != nil {
		return err
	}
	creatorAccount, err := bc.StateDB().GetFormatAccount(matchNft.CreatorAccountIndex)
	if err != nil {
		return err
	}

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.AssetAmount)
	sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance, ffmath.Sub(
		txInfo.BuyOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount)))
	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.CreatorAmount)
	fromAccount.Nonce++

	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	sellOffer = new(big.Int).SetBit(sellOffer, int(e.sellOfferIndex), 1)
	sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized = sellOffer
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.buyOfferIndex), 1)
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer

	// Change new owner.
	matchNft.OwnerAccountIndex = txInfo.BuyOffer.AccountIndex

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(fromAccount.AccountIndex, fromAccount)
	stateCache.SetPendingAccount(buyAccount.AccountIndex, buyAccount)
	stateCache.SetPendingAccount(sellAccount.AccountIndex, sellAccount)
	stateCache.SetPendingAccount(creatorAccount.AccountIndex, creatorAccount)
	stateCache.SetPendingNft(matchNft.NftIndex, matchNft)
	stateCache.SetPendingUpdateGas(txInfo.BuyOffer.AssetId, txInfo.TreasuryAmount)
	stateCache.SetPendingUpdateGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *AtomicMatchExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeAtomicMatch))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
	buf.Write(common2.Uint24ToBytes(txInfo.BuyOffer.OfferId))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
	buf.Write(common2.Uint24ToBytes(txInfo.SellOffer.OfferId))
	buf.Write(common2.Uint40ToBytes(txInfo.BuyOffer.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	packedAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAmountBytes)
	creatorAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.CreatorAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(creatorAmountBytes)
	treasuryAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *AtomicMatchExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.txInfo.GasFeeAssetId
	e.tx.GasFee = e.txInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.txInfo.SellOffer.NftIndex
	e.tx.AssetId = e.txInfo.BuyOffer.AssetId
	e.tx.TxAmount = e.txInfo.BuyOffer.AssetAmount.String()
	return e.BaseExecutor.GetExecutedTx()
}

func (e *AtomicMatchExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	matchNft, err := bc.StateDB().GetNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		return nil, err
	}

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex,
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
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), types.ZeroBigInt).String(),
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
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		AccountName:  buyAccount.AccountName,
		Balance:      buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, ffmath.Neg(txInfo.BuyOffer.AssetAmount), types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           buyAccount.Nonce,
		CollectionNonce: buyAccount.CollectionNonce,
	})
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.AssetAmount)

	// buy offer
	order++
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.buyOfferIndex), 1)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      e.buyOfferAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		AccountName:  buyAccount.AccountName,
		Balance:      buyAccount.AssetInfo[e.buyOfferAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			e.buyOfferAssetId, types.ZeroBigInt, buyOffer).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           buyAccount.Nonce,
		CollectionNonce: buyAccount.CollectionNonce,
	})
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer

	// seller asset A
	order++
	accountOrder++
	sellDeltaAmount := ffmath.Sub(txInfo.SellOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount))
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  sellAccount.AccountName,
		Balance:      sellAccount.AssetInfo[txInfo.SellOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, sellDeltaAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           sellAccount.Nonce,
		CollectionNonce: sellAccount.CollectionNonce,
	})
	sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance, sellDeltaAmount)

	// sell offer
	order++
	sellOffer := sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	sellOffer = new(big.Int).SetBit(sellOffer, int(e.sellOfferIndex), 1)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      e.sellOfferAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		AccountName:  sellAccount.AccountName,
		Balance:      sellAccount.AssetInfo[e.sellOfferAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			e.sellOfferAssetId, types.ZeroBigInt, sellOffer).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           sellAccount.Nonce,
		CollectionNonce: sellAccount.CollectionNonce,
	})
	sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized = sellOffer

	// creator fee
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: matchNft.CreatorAccountIndex,
		AccountName:  creatorAccount.AccountName,
		Balance:      creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.CreatorAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           creatorAccount.Nonce,
		CollectionNonce: creatorAccount.CollectionNonce,
	})
	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.CreatorAmount)

	// nft info
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      matchNft.NftIndex,
		AssetType:    types.NftAssetType,
		AccountIndex: types.NilAccountIndex,
		AccountName:  types.NilAccountName,
		Balance: types.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, matchNft.OwnerAccountIndex,
			matchNft.NftContentHash, matchNft.NftL1TokenId, matchNft.NftL1Address, matchNft.CreatorTreasuryRate, matchNft.CollectionId).String(),
		BalanceDelta: types.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, txInfo.BuyOffer.AccountIndex,
			matchNft.NftContentHash, matchNft.NftL1TokenId, matchNft.NftL1Address, matchNft.CreatorTreasuryRate, matchNft.CollectionId).String(),
		Order:           order,
		AccountOrder:    types.NilAccountOrder,
		Nonce:           0,
		CollectionNonce: 0,
	})

	// gas account asset A - treasury fee
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.TreasuryAmount, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
	})
	gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.TreasuryAmount)

	// gas account asset gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
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
