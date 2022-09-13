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

type AddLiquidityExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.AddLiquidityTxInfo

	newPoolInfo           *types.LiquidityInfo
	lpDeltaForFromAccount *big.Int
}

func NewAddLiquidityExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseAddLiquidityTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &AddLiquidityExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *AddLiquidityExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return errors.New("internal error")
	}
	liquidity := e.bc.StateDB().LiquidityMap[txInfo.PairIndex]

	err = e.BaseExecutor.Prepare(context.WithValue(context.Background(),
		legendTxTypes.TreasuryAccountIndexKey, liquidity.TreasuryAccountIndex))
	if err != nil {
		return err
	}
	// Add the right details to tx info.
	return e.fillTxInfo()
}

func (e *AddLiquidityExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	if txInfo.GasFeeAssetId == txInfo.AssetAId {
		deltaBalance := ffmath.Add(txInfo.AssetAAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetBId].Balance.Cmp(txInfo.AssetBAmount) < 0 {
			return errors.New("invalid asset amount")
		}
	} else if txInfo.GasFeeAssetId == txInfo.AssetBId {
		deltaBalance := ffmath.Add(txInfo.AssetBAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetBId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(txInfo.AssetAAmount) < 0 {
			return errors.New("invalid asset amount")
		}
	} else {
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("invalid gas asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(txInfo.AssetAAmount) < 0 {
			return errors.New("invalid asset amount")
		}
		if fromAccount.AssetInfo[txInfo.AssetBId].Balance.Cmp(txInfo.AssetBAmount) < 0 {
			return errors.New("invalid asset amount")
		}
	}

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return err
	}

	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		return errors.New("invalid liquidity")
	}

	// from account lp
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, txInfo.TreasuryAmount)
	var lpDeltaForFromAccount *big.Int
	if liquidityInfo.AssetA.Cmp(types.ZeroBigInt) == 0 {
		lpDeltaForFromAccount, err = common2.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
		if err != nil {
			logx.Errorf("unable to compute lp delta: %s", err.Error())
			return err
		}
	} else {
		lpDeltaForFromAccount, err = common2.CleanPackedAmount(ffmath.Div(ffmath.Multiply(txInfo.AssetAAmount, poolLp), liquidityInfo.AssetA))
		if err != nil {
			logx.Errorf(" unable to compute lp delta: %s", err.Error())
			return err
		}
	}
	e.lpDeltaForFromAccount = lpDeltaForFromAccount
	e.newPoolInfo = liquidityInfo

	return nil
}

func (e *AddLiquidityExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// apply changes
	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]
	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	treasuryAccount := bc.StateDB().AccountMap[liquidityModel.TreasuryAccountIndex]

	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmount)
	fromAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(fromAccount.AssetInfo[txInfo.PairIndex].LpAmount, e.lpDeltaForFromAccount)
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

func (e *AddLiquidityExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return err
	}

	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		txInfo.LpAmount, err = chain.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("[ComputeEmptyLpAmount] : %v", err)
			return err
		}
	} else {
		txInfo.LpAmount, err = chain.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			return err
		}
	}

	txInfo.AssetAId = liquidityInfo.AssetAId
	txInfo.AssetBId = liquidityInfo.AssetBId

	lpDeltaForTreasuryAccount, err := chain.ComputeSLp(liquidityInfo.AssetA,
		liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		logx.Errorf("[ComputeSLp] err: %v", err)
		return err
	}

	// pool account pool info
	finalPoolA := ffmath.Add(liquidityInfo.AssetA, txInfo.AssetAAmount)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, txInfo.AssetBAmount)

	txInfo.TreasuryAmount = lpDeltaForTreasuryAccount
	txInfo.KLast, err = common2.CleanPackedAmount(ffmath.Multiply(finalPoolA, finalPoolB))
	if err != nil {
		return err
	}

	txInfo.AssetAId = liquidityModel.AssetAId
	txInfo.AssetBId = liquidityModel.AssetBId

	return nil
}

func (e *AddLiquidityExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeAddLiquidity))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetBAmount)
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
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
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

func (e *AddLiquidityExecutor) GetExecutedTx() (*tx.Tx, error) {
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
