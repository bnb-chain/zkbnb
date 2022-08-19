package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type UpdatePairRateExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.UpdatePairRateTxInfo
}

func NewUpdatePairRateExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseUpdatePairRateTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse update pair rate tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &UpdatePairRateExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *UpdatePairRateExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.prepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return err
	}

	return nil
}

func (e *UpdatePairRateExecutor) VerifyInputs() error {
	return nil
}

func (e *UpdatePairRateExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidity := bc.liquidityMap[txInfo.PairIndex]
	liquidity.FeeRate = txInfo.FeeRate
	liquidity.TreasuryAccountIndex = txInfo.TreasuryAccountIndex
	liquidity.TreasuryRate = txInfo.TreasuryRate

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateLiquidityIndexMap[txInfo.PairIndex] = StateCachePending
	return nil
}

func (e *UpdatePairRateExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeUpdatePairRate))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.TreasuryRate)))
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
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *UpdatePairRateExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	return bc.updateLiquidityTree(txInfo.PairIndex)
}

func (e *UpdatePairRateExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *UpdatePairRateExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	liquidity := bc.liquidityMap[txInfo.PairIndex]
	baseLiquidity, err := commonAsset.ConstructLiquidityInfo(
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
	deltaLiquidity := &commonAsset.LiquidityInfo{
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
		AssetId:      txInfo.PairIndex,
		AssetType:    commonAsset.LiquidityAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		Balance:      baseLiquidity.String(),
		BalanceDelta: deltaLiquidity.String(),
		Order:        0,
		AccountOrder: commonConstant.NilAccountOrder,
	}

	return []*tx.TxDetail{txDetail}, nil
}
