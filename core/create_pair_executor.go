package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type CreatePairExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.CreatePairTxInfo
}

func NewCreatePairExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseCreatePairTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse create pair tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &CreatePairExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *CreatePairExecutor) Prepare() error {
	return nil
}

func (e *CreatePairExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	_, err := bc.LiquidityModel.GetLiquidityByPairIndex(txInfo.PairIndex)
	if err != sqlx.ErrNotFound {
		return errors.New("invalid pair index, already registered")
	}

	for index := range bc.stateCache.pendingNewLiquidityIndexMap {
		if txInfo.PairIndex == index {
			return errors.New("invalid pair index, already registered")
		}
	}

	return nil
}

func (e *CreatePairExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	newLiquidity := &liquidity.Liquidity{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               ZeroBigIntString,
		AssetBId:             txInfo.AssetBId,
		AssetB:               ZeroBigIntString,
		LpAmount:             ZeroBigIntString,
		KLast:                ZeroBigIntString,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		FeeRate:              txInfo.FeeRate,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	bc.liquidityMap[txInfo.PairIndex] = newLiquidity

	stateCache := e.bc.stateCache
	stateCache.pendingNewLiquidityIndexMap[txInfo.PairIndex] = StateCachePending
	return nil
}

func (e *CreatePairExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeCreatePair))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetBId)))
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

func (e *CreatePairExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	return bc.updateLiquidityTree(txInfo.PairIndex)
}

func (e *CreatePairExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *CreatePairExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	baseLiquidity := commonAsset.EmptyLiquidityInfo(txInfo.PairIndex)
	deltaLiquidity := &commonAsset.LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               big.NewInt(0),
		AssetBId:             txInfo.AssetBId,
		AssetB:               big.NewInt(0),
		LpAmount:             big.NewInt(0),
		KLast:                big.NewInt(0),
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
