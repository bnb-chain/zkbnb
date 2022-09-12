package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type AtomicMatchExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.AtomicMatchTxInfo

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
	err := e.bc.StateDB().PrepareNft(txInfo.SellOffer.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}
	// Set the right treasury and creator treasury amount.
	matchNft := e.bc.StateDB().NftMap[txInfo.SellOffer.NftIndex]
	txInfo.TreasuryAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(txInfo.SellOffer.TreasuryRate)), big.NewInt(TenThousand))
	txInfo.CreatorAmount = ffmath.Div(ffmath.Multiply(txInfo.SellOffer.AssetAmount, big.NewInt(matchNft.CreatorTreasuryRate)), big.NewInt(TenThousand))
	return e.BaseExecutor.Prepare(context.WithValue(context.Background(), legendTxTypes.CreatorAccountIndexKey, matchNft.CreatorAccountIndex))
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
	if txInfo.AccountIndex == txInfo.BuyOffer.AccountIndex && txInfo.GasFeeAssetId == txInfo.SellOffer.AssetId {
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
	return e.BaseExecutor.GetExecutedTx()
}

func (e *AtomicMatchExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := e.txInfo.Hash(mimc.NewMiMC())
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
