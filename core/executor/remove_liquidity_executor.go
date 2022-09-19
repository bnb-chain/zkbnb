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

type RemoveLiquidityExecutor struct {
	BaseExecutor

	txInfo *txtypes.RemoveLiquidityTxInfo

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
	err = e.bc.StateDB().PrepareAccountsAndAssets(map[int64]map[int64]bool{
		txInfo.FromAccountIndex: {
			txInfo.PairIndex: true,
		},
	})
	if err != nil {
		return err
	}
	err = e.fillTxInfo()
	if err != nil {
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkLiquidityDirty(txInfo.PairIndex)
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId, txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex})
	liquidity, err := e.bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}
	e.MarkAccountAssetsDirty(liquidity.TreasuryAccountIndex, []int64{txInfo.PairIndex})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	return nil
}

func (e *RemoveLiquidityExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("invalid gas asset amount")
	}

	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}
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

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}

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
	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	gasAccount, err := bc.StateDB().GetFormatAccount(txInfo.GasAccountIndex)
	if err != nil {
		return err
	}
	liquidityModel, err := bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return err
	}
	treasuryAccount, err := bc.StateDB().GetFormatAccount(liquidityModel.TreasuryAccountIndex)
	if err != nil {
		return err
	}

	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmountDelta)
	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Sub(fromAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.LpAmount)
	treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.TreasuryAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.StateDB()
	stateCache.SetPendingUpdateAccount(fromAccount.AccountIndex, fromAccount)
	stateCache.SetPendingUpdateAccount(treasuryAccount.AccountIndex, treasuryAccount)
	stateCache.SetPendingUpdateAccount(gasAccount.AccountIndex, gasAccount)
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

func (e *RemoveLiquidityExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	liquidityModel, err := e.bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return nil, err
	}
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: %v", err)
		return nil, err
	}

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex, liquidityInfo.TreasuryAccountIndex})
	if err != nil {
		return nil, err
	}

	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]
	treasuryAccount := copiedAccounts[liquidityInfo.TreasuryAccountIndex]

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
			txInfo.AssetAAmountDelta,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Add(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmountDelta)
	if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(big.NewInt(0)) < 0 {
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
	if fromAccount.AssetInfo[txInfo.AssetBId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient asset b balance")
	}

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
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient gas asset balance")
	}

	// from account lp
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.PairIndex].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.PairIndex,
			types.ZeroBigInt,
			ffmath.Neg(txInfo.LpAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Sub(fromAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.LpAmount)
	if fromAccount.AssetInfo[txInfo.PairIndex].LpAmount.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient lp amount")
	}

	// treasury account
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    types.FungibleAssetType,
		AccountIndex: treasuryAccount.AccountIndex,
		AccountName:  treasuryAccount.AccountName,
		Balance:      treasuryAccount.AssetInfo[txInfo.PairIndex].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.PairIndex, types.ZeroBigInt, txInfo.TreasuryAmount, types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           treasuryAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: treasuryAccount.CollectionNonce,
	})
	treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.TreasuryAmount)

	// pool account info
	liquidity, err := e.bc.StateDB().GetLiquidity(txInfo.PairIndex)
	if err != nil {
		return nil, err
	}
	basePool, err := types.ConstructLiquidityInfo(
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

	finalPoolA := ffmath.Add(liquidityInfo.AssetA, ffmath.Neg(txInfo.AssetAAmountDelta))
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, ffmath.Neg(txInfo.AssetBAmountDelta))
	poolDeltaForToAccount := &types.LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               ffmath.Neg(txInfo.AssetAAmountDelta),
		AssetBId:             txInfo.AssetBId,
		AssetB:               ffmath.Neg(txInfo.AssetBAmountDelta),
		LpAmount:             ffmath.Neg(txInfo.LpAmount),
		KLast:                ffmath.Multiply(finalPoolA, finalPoolB),
		FeeRate:              liquidityInfo.FeeRate,
		TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
		TreasuryRate:         liquidityInfo.TreasuryRate,
	}
	newPool, err := chain.ComputeNewBalance(types.LiquidityAssetType, basePool.String(), poolDeltaForToAccount.String())
	if err != nil {
		return nil, err
	}

	newPoolInfo, err := types.ParseLiquidityInfo(newPool)
	if err != nil {
		return nil, err
	}
	e.newPoolInfo = newPoolInfo
	if newPoolInfo.AssetA.Cmp(big.NewInt(0)) <= 0 ||
		newPoolInfo.AssetB.Cmp(big.NewInt(0)) <= 0 ||
		newPoolInfo.LpAmount.Cmp(big.NewInt(0)) < 0 ||
		newPoolInfo.KLast.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("invalid new pool")
	}

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       types.LiquidityAssetType,
		AccountIndex:    types.NilAccountIndex,
		AccountName:     types.NilAccountName,
		Balance:         basePool.String(),
		BalanceDelta:    poolDeltaForToAccount.String(),
		Order:           order,
		Nonce:           types.NilNonce,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: types.NilCollectionNonce,
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
