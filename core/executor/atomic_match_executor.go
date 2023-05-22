package executor

import (
	"bytes"
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type AtomicMatchExecutor struct {
	BaseExecutor

	TxInfo *txtypes.AtomicMatchTxInfo

	buyOfferAssetId  int64
	buyOfferIndex    int64
	sellOfferAssetId int64
	sellOfferIndex   int64
}

func NewAtomicMatchExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseAtomicMatchTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &AtomicMatchExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewAtomicMatchExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &AtomicMatchExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.AtomicMatchTxInfo),
	}, nil
}

func (e *AtomicMatchExecutor) PreLoadAccountAndNft(accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
	txInfo := e.TxInfo
	accountIndexMap[txInfo.AccountIndex] = true
	accountIndexMap[txInfo.BuyOffer.AccountIndex] = true
	accountIndexMap[txInfo.SellOffer.AccountIndex] = true
	accountIndexMap[txInfo.GasAccountIndex] = true
	accountIndexMap[types.ProtocolAccount] = true
	accountIndexMap[txInfo.BuyOffer.ChannelAccountIndex] = true
	accountIndexMap[txInfo.SellOffer.ChannelAccountIndex] = true

	nftIndexMap[txInfo.SellOffer.NftIndex] = true
}

func (e *AtomicMatchExecutor) Prepare() error {
	txInfo := e.TxInfo

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

	// Set the right creator treasury amount.
	txInfo.RoyaltyAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(matchNft.RoyaltyRate)), big.NewInt(TenThousand))

	// Set the right BuyChanel and SellChanel amount.
	txInfo.BuyChannelAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.BuyOffer.ChannelRate)), big.NewInt(TenThousand))
	txInfo.SellChannelAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.ChannelRate)), big.NewInt(TenThousand))

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.SellOffer.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.BuyOffer.AccountIndex, []int64{txInfo.BuyOffer.AssetId, e.buyOfferAssetId})
	e.MarkAccountAssetsDirty(txInfo.SellOffer.AccountIndex, []int64{txInfo.SellOffer.AssetId, e.sellOfferAssetId})
	e.MarkAccountAssetsDirty(matchNft.CreatorAccountIndex, []int64{txInfo.BuyOffer.AssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.BuyOffer.AssetId, txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(types.ProtocolAccount, []int64{txInfo.BuyOffer.AssetId})
	e.MarkAccountAssetsDirty(txInfo.BuyOffer.ChannelAccountIndex, []int64{txInfo.BuyOffer.AssetId})
	e.MarkAccountAssetsDirty(txInfo.SellOffer.ChannelAccountIndex, []int64{txInfo.SellOffer.AssetId})
	return e.BaseExecutor.Prepare()
}

func (e *AtomicMatchExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	bc := e.bc
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	protocolRate, err := bc.StateDB().GetProtocolRateFromRedisCache()
	if err != nil {
		return err
	}

	if protocolRate != txInfo.BuyOffer.ProtocolRate {
		return types.AppErrInvalidPlatformRate
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
		txInfo.SellOffer.AssetAmount.String() != txInfo.BuyOffer.AssetAmount.String() {
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
	_, err = bc.StateDB().GetFormatAccount(txInfo.BuyOffer.ChannelAccountIndex)
	if err != nil {
		return err
	}
	_, err = bc.StateDB().GetFormatAccount(txInfo.SellOffer.ChannelAccountIndex)
	if err != nil {
		return err
	}
	matchNft, err := e.bc.StateDB().PrepareNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		return err
	}
	if matchNft.RoyaltyRate != txInfo.BuyOffer.RoyaltyRate {
		return types.AppErrInvalidRoyaltyRate
	}

	// Check sender's gas balance and buyer's asset balance.
	if txInfo.AccountIndex == txInfo.BuyOffer.AccountIndex && txInfo.GasFeeAssetId == txInfo.SellOffer.AssetId {
		totalBalance := ffmath.Add(ffmath.Add(ffmath.Add(ffmath.Add(
			txInfo.BuyOffer.AssetAmount, txInfo.BuyOffer.ProtocolAmount),
			txInfo.RoyaltyAmount),
			txInfo.BuyChannelAmount),
			txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(totalBalance) < 0 {
			return types.AppErrBuyerBalanceNotEnough
		}
	} else {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			if txInfo.AccountIndex == txInfo.BuyOffer.AccountIndex {
				return types.AppErrBuyerBalanceNotEnough
			} else if txInfo.AccountIndex == txInfo.SellOffer.AccountIndex {
				return types.AppErrSellerBalanceNotEnough
			} else {
				return types.AppErrCommitNotEnough
			}
		}
		buyDeltaAmount := ffmath.Add(ffmath.Add(ffmath.Add(
			txInfo.BuyOffer.AssetAmount, txInfo.BuyOffer.ProtocolAmount),
			txInfo.RoyaltyAmount),
			txInfo.BuyChannelAmount)

		if buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance.Cmp(buyDeltaAmount) < 0 {
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

	//committer skip check
	if !skipSigChk {
		// Verify l1 signature.
		if txInfo.SellOffer.GetL1AddressBySignature() != common.HexToAddress(sellAccount.L1Address) {
			return types.AppErrFailToL1Signature
		}
		if txInfo.BuyOffer.GetL1AddressBySignature() != common.HexToAddress(buyAccount.L1Address) {
			return types.AppErrFailToL1Signature
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
	}
	return nil
}

func (e *AtomicMatchExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

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
	buyChannelAccount, err := bc.StateDB().GetFormatAccount(txInfo.BuyOffer.ChannelAccountIndex)
	if err != nil {
		return err
	}
	sellChannelAccount, err := bc.StateDB().GetFormatAccount(txInfo.SellOffer.ChannelAccountIndex)
	if err != nil {
		return err
	}
	protocolAccount, err := bc.StateDB().GetFormatAccount(types.ProtocolAccount)
	if err != nil {
		return err
	}

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance,
		ffmath.Add(ffmath.Add(ffmath.Add(
			txInfo.BuyOffer.AssetAmount, txInfo.BuyOffer.ProtocolAmount),
			txInfo.RoyaltyAmount),
			txInfo.BuyChannelAmount))
	sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance,
		ffmath.Sub(txInfo.BuyOffer.AssetAmount, txInfo.SellChannelAmount))

	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.RoyaltyAmount)
	buyChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(buyChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyChannelAmount)
	sellChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(sellChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.SellChannelAmount)
	protocolAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(protocolAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.ProtocolAmount)

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
	stateCache.SetPendingAccount(buyChannelAccount.AccountIndex, buyChannelAccount)
	stateCache.SetPendingAccount(sellChannelAccount.AccountIndex, sellChannelAccount)
	stateCache.SetPendingNft(matchNft.NftIndex, matchNft)
	stateCache.SetPendingAccount(types.ProtocolAccount, protocolAccount)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *AtomicMatchExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeAtomicMatch))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
	buf.Write(common2.Uint24ToBytes(txInfo.BuyOffer.OfferId))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
	buf.Write(common2.Uint24ToBytes(txInfo.SellOffer.OfferId))
	buf.Write(common2.Uint40ToBytes(txInfo.BuyOffer.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	packedAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAmountBytes)

	royaltyAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.RoyaltyAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(royaltyAmountBytes)

	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)

	protocolAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.BuyOffer.ProtocolAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(protocolAmountBytes)

	buf.Write(common2.Uint32ToBytes(uint32(txInfo.BuyOffer.ChannelAccountIndex)))
	buyChanelAmount, err := common2.AmountToPackedAmountBytes(txInfo.BuyChannelAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(buyChanelAmount)

	buf.Write(common2.Uint32ToBytes(uint32(txInfo.SellOffer.ChannelAccountIndex)))
	sellChanelAmount, err := common2.AmountToPackedAmountBytes(txInfo.SellChannelAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(sellChanelAmount)

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *AtomicMatchExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.TxInfo.SellOffer.NftIndex
	e.tx.AssetId = e.TxInfo.BuyOffer.AssetId
	e.tx.TxAmount = e.TxInfo.BuyOffer.AssetAmount.String()
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *AtomicMatchExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.TxInfo
	matchNft, err := bc.StateDB().GetNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		return nil, err
	}

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex,
		txInfo.SellOffer.AccountIndex, txInfo.BuyOffer.AccountIndex, matchNft.CreatorAccountIndex,
		txInfo.BuyOffer.ChannelAccountIndex, txInfo.SellOffer.ChannelAccountIndex, types.ProtocolAccount})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.AccountIndex]
	buyAccount := copiedAccounts[txInfo.BuyOffer.AccountIndex]
	sellAccount := copiedAccounts[txInfo.SellOffer.AccountIndex]
	creatorAccount := copiedAccounts[matchNft.CreatorAccountIndex]
	buyChannelAccount := copiedAccounts[txInfo.BuyOffer.ChannelAccountIndex]
	sellChannelAccount := copiedAccounts[txInfo.SellOffer.ChannelAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]
	protocolAccount := copiedAccounts[types.ProtocolAccount]

	txDetails := make([]*tx.TxDetail, 0, 11)

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
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	// buyer asset A
	order++
	accountOrder++
	buyDeltaAmount := ffmath.Add(ffmath.Add(ffmath.Add(
		txInfo.BuyOffer.AssetAmount, txInfo.BuyOffer.ProtocolAmount),
		txInfo.RoyaltyAmount),
		txInfo.BuyChannelAmount)

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		L1Address:    buyAccount.L1Address,
		Balance:      buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, ffmath.Neg(buyDeltaAmount), types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           buyAccount.Nonce,
		CollectionNonce: buyAccount.CollectionNonce,
		PublicKey:       buyAccount.PublicKey,
	})
	buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Sub(buyAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, buyDeltaAmount)

	// buy offer
	order++
	buyOffer := buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.buyOfferIndex), 1)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      e.buyOfferAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		L1Address:    buyAccount.L1Address,
		Balance:      buyAccount.AssetInfo[e.buyOfferAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			e.buyOfferAssetId, types.ZeroBigInt, buyOffer).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           buyAccount.Nonce,
		CollectionNonce: buyAccount.CollectionNonce,
		PublicKey:       buyAccount.PublicKey,
	})
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer

	// seller asset A
	order++
	accountOrder++
	sellDeltaAmount := ffmath.Sub(txInfo.SellOffer.AssetAmount, txInfo.SellChannelAmount)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		L1Address:    sellAccount.L1Address,
		Balance:      sellAccount.AssetInfo[txInfo.SellOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, sellDeltaAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           sellAccount.Nonce,
		CollectionNonce: sellAccount.CollectionNonce,
		PublicKey:       sellAccount.PublicKey,
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
		L1Address:    sellAccount.L1Address,
		Balance:      sellAccount.AssetInfo[e.sellOfferAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			e.sellOfferAssetId, types.ZeroBigInt, sellOffer).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           sellAccount.Nonce,
		CollectionNonce: sellAccount.CollectionNonce,
		PublicKey:       sellAccount.PublicKey,
	})
	sellAccount.AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized = sellOffer

	// creator fee
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: matchNft.CreatorAccountIndex,
		L1Address:    creatorAccount.L1Address,
		Balance:      creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.RoyaltyAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           creatorAccount.Nonce,
		CollectionNonce: creatorAccount.CollectionNonce,
		PublicKey:       creatorAccount.PublicKey,
	})
	creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(creatorAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.RoyaltyAmount)

	// nft info
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      matchNft.NftIndex,
		AssetType:    types.NftAssetType,
		AccountIndex: types.NilAccountIndex,
		L1Address:    types.NilL1Address,
		Balance: types.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, matchNft.OwnerAccountIndex,
			matchNft.NftContentHash, matchNft.RoyaltyRate, matchNft.CollectionId, matchNft.NftContentType).String(),
		BalanceDelta: types.ConstructNftInfo(matchNft.NftIndex, matchNft.CreatorAccountIndex, txInfo.BuyOffer.AccountIndex,
			matchNft.NftContentHash, matchNft.RoyaltyRate, matchNft.CollectionId, matchNft.NftContentType).String(),
		Order:           order,
		AccountOrder:    types.NilAccountOrder,
		Nonce:           0,
		CollectionNonce: 0,
		PublicKey:       types.EmptyPk,
	})

	// buyChannelAccount
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.BuyOffer.ChannelAccountIndex,
		L1Address:    buyChannelAccount.L1Address,
		Balance:      buyChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.BuyChannelAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           buyChannelAccount.Nonce,
		CollectionNonce: buyChannelAccount.CollectionNonce,
		PublicKey:       buyChannelAccount.PublicKey,
	})
	buyChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(buyChannelAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyChannelAmount)

	// sellChannelAccount
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.SellOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.SellOffer.ChannelAccountIndex,
		L1Address:    sellChannelAccount.L1Address,
		Balance:      sellChannelAccount.AssetInfo[txInfo.SellOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.SellOffer.AssetId, txInfo.SellChannelAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           sellChannelAccount.Nonce,
		CollectionNonce: sellChannelAccount.CollectionNonce,
		PublicKey:       sellChannelAccount.PublicKey,
	})
	sellChannelAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellChannelAccount.AssetInfo[txInfo.SellOffer.AssetId].Balance, txInfo.SellChannelAmount)

	// protocol Account amount
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: protocolAccount.AccountIndex,
		L1Address:    protocolAccount.L1Address,
		Balance:      protocolAccount.AssetInfo[txInfo.BuyOffer.AssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.BuyOffer.AssetId, txInfo.BuyOffer.ProtocolAmount, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           protocolAccount.Nonce,
		CollectionNonce: protocolAccount.CollectionNonce,
		PublicKey:       protocolAccount.PublicKey,
	})
	protocolAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance = ffmath.Add(protocolAccount.AssetInfo[txInfo.BuyOffer.AssetId].Balance, txInfo.BuyOffer.ProtocolAmount)

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

func (e *AtomicMatchExecutor) Finalize() error {
	metrics.ProtocolFeeRevenueCounter.Add(common2.GetFeeFromWei(e.TxInfo.BuyOffer.ProtocolAmount))
	metrics.TotalRevenueCounter.Add(common2.GetFeeFromWei(e.TxInfo.BuyOffer.ProtocolAmount))
	return nil
}
