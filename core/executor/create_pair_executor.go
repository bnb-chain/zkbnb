package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
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
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *CreatePairExecutor) Prepare() error {
	// Mark the tree states that would be affected in this executor.
	return e.BaseExecutor.Prepare(context.Background())
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
	return e.BaseExecutor.ApplyTransaction()
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
