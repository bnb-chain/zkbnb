package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type RemoveLiquidityExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.RemoveLiquidityTxInfo

	newPoolInfo *types.LiquidityInfo
}

func NewRemoveLiquidityExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseRemoveLiquidityTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &RemoveLiquidityExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *RemoveLiquidityExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return err
	}

	liquidity := e.bc.StateDB().LiquidityMap[txInfo.PairIndex]

	err = e.BaseExecutor.Prepare(context.WithValue(context.Background(),
		legendTxTypes.TreasuryAccountIndexKey, liquidity.TreasuryAccountIndex))
	if err != nil {
		return err
	}

	return e.fillTxInfo()
}

func (e *RemoveLiquidityExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("invalid gas asset amount")
	}

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return err
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.LpAmount == nil || liquidityInfo.LpAmount.Cmp(big.NewInt(0)) == 0 {
		return errors.New("invalid pool liquidity")
	}

	return nil
}

func (e *RemoveLiquidityExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return err
	}

	assetAAmount, assetBAmount, err := chain.ComputeRemoveLiquidityAmount(liquidityInfo, txInfo.LpAmount)
	if err != nil {
		return err
	}

	if assetAAmount.Cmp(txInfo.AssetAMinAmount) < 0 || assetBAmount.Cmp(txInfo.AssetBMinAmount) < 0 {
		return errors.New("invalid asset min amount")
	}

	if fromAccount.AssetInfo[txInfo.PairIndex].LpAmount.Cmp(txInfo.LpAmount) < 0 {
		return errors.New("invalid lp amount")
	}

	txInfo.AssetAAmountDelta = assetAAmount
	txInfo.AssetBAmountDelta = assetBAmount
	txInfo.AssetAId = liquidityInfo.AssetAId
	txInfo.AssetBId = liquidityInfo.AssetBId

	poolAssetADelta := ffmath.Neg(txInfo.AssetAAmountDelta)
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
	finalPoolA := ffmath.Add(liquidityInfo.AssetA, poolAssetADelta)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, poolAssetBDelta)
	lpDeltaForTreasuryAccount, err := chain.ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		return err
	}

	// set tx info
	txInfo.KLast, err = common2.CleanPackedAmount(ffmath.Multiply(finalPoolA, finalPoolB))
	if err != nil {
		return err
	}
	txInfo.TreasuryAmount = lpDeltaForTreasuryAccount

	return nil
}

func (e *RemoveLiquidityExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// apply changes
	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]
	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	treasuryAccount := bc.StateDB().AccountMap[liquidityModel.TreasuryAccountIndex]

	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmountDelta)
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Sub(fromAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.LpAmount)
	treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.TreasuryAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

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
	stateCache.PendingUpdateAccountIndexMap[treasuryAccount.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateLiquidityIndexMap[txInfo.PairIndex] = statedb.StateCachePending
	return e.BaseExecutor.ApplyTransaction()
}

func (e *RemoveLiquidityExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeRemoveLiquidity))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetAAmountDelta)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := common2.AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(KLastBytes)
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
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

func (e *RemoveLiquidityExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.txInfo.GasFeeAssetId
	e.tx.GasFee = e.txInfo.GasFeeAssetAmount.String()
	e.tx.PairIndex = e.txInfo.PairIndex
	e.tx.TxAmount = e.txInfo.LpAmount.String()
	return e.BaseExecutor.GetExecutedTx()
}

