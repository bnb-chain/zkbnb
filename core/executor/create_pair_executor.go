package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

type CreatePairExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.CreatePairTxInfo
}

func NewCreatePairExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseCreatePairTxInfo(tx.TxInfo)
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

	_, err := bc.DB().LiquidityModel.GetLiquidityByIndex(txInfo.PairIndex)
	if err != sqlx.ErrNotFound {
		return errors.New("invalid pair index, already registered")
	}

	for index := range bc.StateDB().PendingNewLiquidityIndexMap {
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
		AssetA:               types.ZeroBigIntString,
		AssetBId:             txInfo.AssetBId,
		AssetB:               types.ZeroBigIntString,
		LpAmount:             types.ZeroBigIntString,
		KLast:                types.ZeroBigIntString,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		FeeRate:              txInfo.FeeRate,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	bc.StateDB().LiquidityMap[txInfo.PairIndex] = newLiquidity

	stateCache := e.bc.StateDB()
	stateCache.PendingNewLiquidityIndexMap[txInfo.PairIndex] = statedb.StateCachePending
	return nil
}

func (e *CreatePairExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeCreatePair))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(common.Uint16ToBytes(uint16(txInfo.AssetBId)))
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

func (e *CreatePairExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	return bc.StateDB().UpdateLiquidityTree(txInfo.PairIndex)
}

func (e *CreatePairExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.PairIndex = e.txInfo.PairIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *CreatePairExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	baseLiquidity := types.EmptyLiquidityInfo(txInfo.PairIndex)
	deltaLiquidity := &types.LiquidityInfo{
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

func (e *CreatePairExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	return nil, nil
}
