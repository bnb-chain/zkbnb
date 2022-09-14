package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type UpdatePairRateExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.UpdatePairRateTxInfo
}

func NewUpdatePairRateExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseUpdatePairRateTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse update pair rate tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &UpdatePairRateExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *UpdatePairRateExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkLiquidityDirty(txInfo.PairIndex)
	return e.BaseExecutor.Prepare()
}

func (e *UpdatePairRateExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo
	liquidity, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}

	if liquidity.FeeRate == txInfo.FeeRate &&
		liquidity.TreasuryAccountIndex == txInfo.TreasuryAccountIndex &&
		liquidity.TreasuryRate == txInfo.TreasuryRate {
		return errors.New("invalid update, the same to old")
	}

	return nil
}

func (e *UpdatePairRateExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidity, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}
	liquidity.FeeRate = txInfo.FeeRate
	liquidity.TreasuryAccountIndex = txInfo.TreasuryAccountIndex
	liquidity.TreasuryRate = txInfo.TreasuryRate

	stateCache := e.bc.StateDB()
	stateCache.SetPendingUpdateLiquidity(txInfo.PairIndex, liquidity)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *UpdatePairRateExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeUpdatePairRate))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(common.Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.TreasuryRate)))
	chunk := common.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *UpdatePairRateExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.PairIndex = e.txInfo.PairIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *UpdatePairRateExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	liquidity, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return nil, err
	}
	baseLiquidity, err := types.ConstructLiquidityInfo(
		liquidity.PairIndex,
		liquidity.AssetAId,
		liquidity.AssetA,
		liquidity.AssetBId,
		liquidity.AssetB,
		liquidity.LpAmount,
		liquidity.KLast,
		liquidity.FeeRate,
		liquidity.TreasuryAccountIndex,
		liquidity.TreasuryRate,
	)
	if err != nil {
		return nil, err
	}
	deltaLiquidity := &types.LiquidityInfo{
		PairIndex:            baseLiquidity.PairIndex,
		AssetAId:             baseLiquidity.AssetAId,
		AssetA:               big.NewInt(0),
		AssetBId:             baseLiquidity.AssetBId,
		AssetB:               big.NewInt(0),
		LpAmount:             big.NewInt(0),
		KLast:                baseLiquidity.KLast,
		FeeRate:              txInfo.FeeRate,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		TreasuryRate:         txInfo.TreasuryRate,
	}

	txDetail := &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       types.LiquidityAssetType,
		AccountIndex:    types.NilAccountIndex,
		AccountName:     types.NilAccountName,
		Balance:         baseLiquidity.String(),
		BalanceDelta:    deltaLiquidity.String(),
		Order:           0,
		AccountOrder:    types.NilAccountOrder,
		Nonce:           types.NilNonce,
		CollectionNonce: types.NilCollectionNonce,
	}

	return []*tx.TxDetail{txDetail}, nil
}
