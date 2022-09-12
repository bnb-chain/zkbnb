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
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type SwapExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.SwapTxInfo

	newPoolInfo *types.LiquidityInfo
}

func NewSwapExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseSwapTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &SwapExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *SwapExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return errors.New("internal error")
	}

	err = e.BaseExecutor.Prepare(context.Background())
	if err != nil {
		return err
	}

	return e.fillTxInfo()
}

func (e *SwapExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	if txInfo.GasFeeAssetId != txInfo.AssetAId {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("invalid gas asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(txInfo.AssetAAmount) < 0 {
			return errors.New("invalid asset amount")
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
	}

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return errors.New("internal error")
	}
	if !((liquidityModel.AssetAId == txInfo.AssetAId && liquidityModel.AssetBId == txInfo.AssetBId) ||
		(liquidityModel.AssetAId == txInfo.AssetBId && liquidityModel.AssetBId == txInfo.AssetAId)) {
		return errors.New("invalid asset ids")
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 {
		return errors.New("liquidity is empty")
	}

	return nil
}

func constructLiquidityInfo(liquidity *liquidity.Liquidity) (*types.LiquidityInfo, error) {
	return types.ConstructLiquidityInfo(
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
}

func (e *SwapExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// apply changes
	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]

	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	bc.StateDB().LiquidityMap[txInfo.PairIndex] = &liquidity.Liquidity{
		Model:                liquidityModel.Model,
		PairIndex:            e.newPoolInfo.PairIndex,
		AssetAId:             liquidityModel.AssetAId,
		AssetA:               e.newPoolInfo.AssetA.String(),
		AssetBId:             liquidityModel.AssetBId,
		AssetB:               e.newPoolInfo.AssetB.String(),
		LpAmount:             e.newPoolInfo.LpAmount.String(),
		KLast:                e.newPoolInfo.KLast.String(),
		FeeRate:              e.newPoolInfo.FeeRate,
		TreasuryAccountIndex: e.newPoolInfo.TreasuryAccountIndex,
		TreasuryRate:         e.newPoolInfo.TreasuryRate,
	}

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateLiquidityIndexMap[txInfo.PairIndex] = statedb.StateCachePending
	return e.BaseExecutor.ApplyTransaction()
}

func (e *SwapExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return err
	}

	// add details to tx info
	var toDelta *big.Int
	if liquidityInfo.AssetAId == txInfo.AssetAId && liquidityInfo.AssetBId == txInfo.AssetBId {
		toDelta, _, err = chain.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
		if err != nil {
			return err
		}
	} else if liquidityInfo.AssetAId == txInfo.AssetBId && liquidityInfo.AssetBId == txInfo.AssetAId {
		toDelta, _, err = chain.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetBId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
		if err != nil {
			return err
		}
	}

	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		return errors.New("invalid AssetBMinAmount")
	}
	txInfo.AssetBAmountDelta = toDelta

	return nil
}

func (e *SwapExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeSwap))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetAId)))
	packedAssetAAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetAAmountBytes)
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetBId)))
	packedAssetBAmountDeltaBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *SwapExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *SwapExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
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
		PairIndex:     e.txInfo.PairIndex,
		AssetId:       types.NilAssetId,
		TxAmount:      e.txInfo.AssetAAmount.String(),
		Memo:          "",
		AccountIndex:  e.txInfo.FromAccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
