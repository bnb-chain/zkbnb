package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
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
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *RemoveLiquidityExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return err
	}

	liquidityModel := e.bc.StateDB().LiquidityMap[txInfo.PairIndex]

	accounts := []int64{txInfo.FromAccountIndex, liquidityModel.TreasuryAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{liquidityModel.AssetAId, liquidityModel.AssetBId, txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex, txInfo.GasFeeAssetId}
	err = e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	err = e.fillTxInfo()
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

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("invalid gas asset amount")
	}

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
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
		logx.Errorf("construct liquidity info error, err: ", err.Error())
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
	fromAccountInfo := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	gasAccountInfo := bc.StateDB().AccountMap[txInfo.GasAccountIndex]
	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]
	treasuryAccount := bc.StateDB().AccountMap[liquidityModel.TreasuryAccountIndex]

	fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance = ffmath.Add(fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmountDelta)
	fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccountInfo.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Sub(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.LpAmount)
	treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Sub(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.TreasuryAmount)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccountInfo.Nonce++

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
	return nil
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

func (e *RemoveLiquidityExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]

	accounts := []int64{txInfo.FromAccountIndex, liquidityModel.TreasuryAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex, txInfo.GasFeeAssetId}

	err := bc.StateDB().UpdateAccountTree(accounts, assets)
	if err != nil {
		return err
	}

	err = bc.StateDB().UpdateLiquidityTree(txInfo.PairIndex)
	if err != nil {
		return err
	}

	return nil
}

func (e *RemoveLiquidityExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *RemoveLiquidityExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	liquidityModel := e.bc.StateDB().LiquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
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
		AccountName:  treasuryAccount.AccountNameHash,
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
	basePool, err := types.ConstructLiquidityInfo(
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].PairIndex,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].AssetAId,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].AssetA,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].AssetBId,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].AssetB,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].LpAmount,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].KLast,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].FeeRate,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].TreasuryAccountIndex,
		e.bc.StateDB().LiquidityMap[txInfo.PairIndex].TreasuryRate,
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

func (e *RemoveLiquidityExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := legendTxTypes.ComputeRemoveLiquidityMsgHash(e.txInfo, mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      types.NilTxNftIndex,
		PairIndex:     e.txInfo.PairIndex,
		AssetId:       types.NilAssetId,
		TxAmount:      "",
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
