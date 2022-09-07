package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
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

func NewAtomicMatchExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseAtomicMatchTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &AtomicMatchExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *AtomicMatchExecutor) Prepare() error {
	txInfo := e.txInfo

	e.buyOfferAssetId = txInfo.BuyOffer.OfferId / OfferPerAsset
	e.buyOfferIndex = txInfo.BuyOffer.OfferId % OfferPerAsset
	e.sellOfferAssetId = txInfo.SellOffer.OfferId / OfferPerAsset
	e.sellOfferIndex = txInfo.SellOffer.OfferId % OfferPerAsset

	// Prepare seller's asset and nft, if the buyer's asset or nft isn't the same, it will be failed in the verify step.
	err := e.bc.StateDB().PrepareNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}

	matchNft := e.bc.StateDB().NftMap[txInfo.SellOffer.NftIndex]
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
	err = e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	// Set the right treasury and creator treasury amount.
	txInfo.TreasuryAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.TreasuryRate)), big.NewInt(TenThousand))
	txInfo.CreatorAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(matchNft.CreatorTreasuryRate)), big.NewInt(TenThousand))

	return nil
}

func (e *AtomicMatchExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	if txInfo.BuyOffer.Type != types.BuyOfferType ||
		txInfo.SellOffer.Type != types.SellOfferType {
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

	// Check offer expired time.
	if err := e.bc.VerifyExpiredAt(txInfo.BuyOffer.ExpiredAt); err != nil {
		return errors.New("invalid BuyOffer.ExpiredAt")
	}
	if err := e.bc.VerifyExpiredAt(txInfo.SellOffer.ExpiredAt); err != nil {
		return errors.New("invalid SellOffer.ExpiredAt")
	}

	fromAccount := bc.StateDB().AccountMap[txInfo.AccountIndex]
	buyAccount := bc.StateDB().AccountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.StateDB().AccountMap[txInfo.SellOffer.AccountIndex]

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
	sellOffer := bc.StateDB().AccountMap[txInfo.SellOffer.AccountIndex].AssetInfo[e.sellOfferAssetId].OfferCanceledOrFinalized
	if sellOffer.Bit(int(e.sellOfferIndex)) == 1 {
		return errors.New("sell offer canceled or finalized")
	}
	buyOffer := bc.StateDB().AccountMap[txInfo.BuyOffer.AccountIndex].AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized
	if buyOffer.Bit(int(e.buyOfferIndex)) == 1 {
		return errors.New("buy offer canceled or finalized")
	}

	// Check the seller is the owner of the nft.
	if bc.StateDB().NftMap[txInfo.SellOffer.NftIndex].OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		return errors.New("seller is not owner")
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
	matchNft := bc.StateDB().NftMap[txInfo.SellOffer.NftIndex]
	fromAccount := bc.StateDB().AccountMap[txInfo.AccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]
	buyAccount := bc.StateDB().AccountMap[txInfo.BuyOffer.AccountIndex]
	sellAccount := bc.StateDB().AccountMap[txInfo.SellOffer.AccountIndex]
	creatorAccount := bc.StateDB().AccountMap[matchNft.CreatorAccountIndex]

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
	buyOffer = new(big.Int).SetBit(buyOffer, int(e.buyOfferIndex), 1)
	buyAccount.AssetInfo[e.buyOfferAssetId].OfferCanceledOrFinalized = buyOffer

	// Change new owner.
	matchNft.OwnerAccountIndex = txInfo.BuyOffer.AccountIndex

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.BuyOffer.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.SellOffer.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[matchNft.CreatorAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateNftIndexMap[txInfo.SellOffer.NftIndex] = statedb.StateCachePending

	return nil
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

func (e *AtomicMatchExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	matchNft := e.bc.StateDB().NftMap[txInfo.SellOffer.NftIndex]
	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex, txInfo.SellOffer.AccountIndex, matchNft.CreatorAccountIndex}
	if !e.isFromBuyer {
		accounts = append(accounts, txInfo.BuyOffer.AccountIndex)
	}
	assets := []int64{txInfo.GasFeeAssetId, e.buyOfferAssetId, e.sellOfferAssetId}
	if !e.isAssetGas {
		assets = append(assets, txInfo.SellOffer.AssetId)
	}
	err := e.bc.StateDB().UpdateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.StateDB().UpdateNftTree(txInfo.SellOffer.NftIndex)
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
	matchNft := bc.StateDB().NftMap[txInfo.SellOffer.NftIndex]

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
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), types.ZeroBigInt, types.ZeroBigInt).String(),
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
			txInfo.BuyOffer.AssetId, ffmath.Neg(txInfo.BuyOffer.AssetAmount), types.ZeroBigInt, types.ZeroBigInt,
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
			e.buyOfferAssetId, types.ZeroBigInt, types.ZeroBigInt, buyOffer).String(),
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
			txInfo.SellOffer.AssetId, sellDeltaAmount, types.ZeroBigInt, types.ZeroBigInt,
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
			e.sellOfferAssetId, types.ZeroBigInt, types.ZeroBigInt, sellOffer).String(),
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
			txInfo.BuyOffer.AssetId, txInfo.CreatorAmount, types.ZeroBigInt, types.ZeroBigInt,
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
			txInfo.BuyOffer.AssetId, txInfo.TreasuryAmount, types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
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
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, types.ZeroBigInt, types.ZeroBigInt).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	return txDetails, nil
}

func (e *AtomicMatchExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := legendTxTypes.ComputeAtomicMatchMsgHash(e.txInfo, mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      types.NilNftIndex,
		PairIndex:     types.NilPairIndex,
		AssetId:       e.txInfo.BuyOffer.AssetId,
		TxAmount:      e.txInfo.BuyOffer.AssetAmount.String(),
		Memo:          "",
		AccountIndex:  e.txInfo.AccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
