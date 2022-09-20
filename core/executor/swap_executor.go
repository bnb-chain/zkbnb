package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"

	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type SwapExecutor struct {
	BaseExecutor

	txInfo *txtypes.SwapTxInfo

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

	err = e.fillTxInfo()
	if err != nil {
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkLiquidityDirty(txInfo.PairIndex)
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetAId, txInfo.AssetBId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	return nil
}

func (e *SwapExecutor) VerifyInputs(skipGasAmtChk bool) error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk)
	if err != nil {
		return err
	}

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
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

	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return errors.New("internal error")
	}
	if !((liquidityModel.AssetAId == txInfo.AssetAId && liquidityModel.AssetBId == txInfo.AssetBId) ||
		(liquidityModel.AssetAId == txInfo.AssetBId && liquidityModel.AssetBId == txInfo.AssetAId)) {
		return errors.New("invalid asset ids")
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(types.ZeroBigInt) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(types.ZeroBigInt) == 0 {
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
	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	gasAccount, err := bc.StateDB().GetFormatAccount(txInfo.GasAccountIndex)
	if err != nil {
		return err
	}

	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}

	stateCache := e.bc.StateDB()
	stateCache.SetPendingUpdateAccount(txInfo.FromAccountIndex, fromAccount)
	stateCache.SetPendingUpdateAccount(txInfo.GasAccountIndex, gasAccount)
	stateCache.SetPendingUpdateLiquidity(txInfo.PairIndex, &liquidity.Liquidity{
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
	})
	return e.BaseExecutor.ApplyTransaction()
}

func (e *SwapExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}

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
	e.tx.GasFeeAssetId = e.txInfo.GasFeeAssetId
	e.tx.GasFee = e.txInfo.GasFeeAssetAmount.String()
	e.tx.PairIndex = e.txInfo.PairIndex
	e.tx.TxAmount = e.txInfo.AssetAAmount.String()
	return e.BaseExecutor.GetExecutedTx()
}

func (e *SwapExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}

	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]
	liquidityModel, err := e.bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return nil, err
	}
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return nil, err
	}

	txDetails := make([]*tx.TxDetail, 0, 4)
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetAId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetAId,
			ffmath.Neg(txInfo.AssetAAmount),
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient asset a balance")
	}

	// from account asset B
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetBId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.AssetBId,
			txInfo.AssetBAmountDelta,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)

	// from account asset gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// pool info
	var poolDelta *types.LiquidityInfo
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
	if txInfo.AssetAId == liquidityInfo.AssetAId {
		poolDelta = &types.LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetAId,
			AssetA:               txInfo.AssetAAmount,
			AssetBId:             txInfo.AssetBId,
			AssetB:               poolAssetBDelta,
			LpAmount:             types.ZeroBigInt,
			KLast:                types.ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	} else if txInfo.AssetAId == liquidityInfo.AssetBId {
		poolDelta = &types.LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetBId,
			AssetA:               poolAssetBDelta,
			AssetBId:             txInfo.AssetAId,
			AssetB:               txInfo.AssetAAmount,
			LpAmount:             types.ZeroBigInt,
			KLast:                types.ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	}

	newPool, err := chain.ComputeNewBalance(
		types.LiquidityAssetType, liquidityInfo.String(), poolDelta.String())
	if err != nil {
		return nil, err
	}

	nPoolInfo, err := types.ParseLiquidityInfo(newPool)
	if err != nil {
		return nil, err
	}
	e.newPoolInfo = nPoolInfo

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       types.LiquidityAssetType,
		AccountIndex:    types.NilAccountIndex,
		AccountName:     types.NilAccountName,
		Balance:         liquidityInfo.String(),
		BalanceDelta:    poolDelta.String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: 0,
	})

	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails, nil
}
