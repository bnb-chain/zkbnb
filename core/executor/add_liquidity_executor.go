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
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *AddLiquidityExecutor) Prepare() error {
	txInfo := e.txInfo

	err := e.bc.StateDB().PrepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return errors.New("internal error")
	}

	liquidityModel := e.bc.StateDB().LiquidityMap[txInfo.PairIndex]

	accounts := []int64{txInfo.FromAccountIndex, liquidityModel.TreasuryAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{liquidityModel.AssetAId, liquidityModel.AssetBId, txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex, txInfo.GasFeeAssetId}
	err = e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	// add details to tx info
	err = e.fillTxInfo()
	if err != nil {
		return err
	}

	return nil
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
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return err
	}

	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		return errors.New("invalid liquidity")
	}

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
	return nil
}

func (e *AddLiquidityExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.StateDB().LiquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
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

func (e *AddLiquidityExecutor) UpdateTrees() error {
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

func (e *AddLiquidityExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *AddLiquidityExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
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
			ffmath.Neg(txInfo.AssetBAmount),
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})

	fromAccount.AssetInfo[txInfo.AssetBId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmount)
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
		return nil, errors.New("insufficient gas fee balance")
	}

	// from account lp
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, txInfo.TreasuryAmount)
	var lpDeltaForFromAccount *big.Int
	if liquidityInfo.AssetA.Cmp(types.ZeroBigInt) == 0 {
		lpDeltaForFromAccount, err = common2.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
		if err != nil {
			logx.Errorf("unable to compute lp delta: %s", err.Error())
			return nil, err
		}
	} else {
		lpDeltaForFromAccount, err = common2.CleanPackedAmount(ffmath.Div(ffmath.Multiply(txInfo.AssetAAmount, poolLp), liquidityInfo.AssetA))
		if err != nil {
			logx.Errorf(" unable to compute lp delta: %s", err.Error())
			return nil, err
		}
	}

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
			lpDeltaForFromAccount,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	e.lpDeltaForFromAccount = lpDeltaForFromAccount
	fromAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(fromAccount.AssetInfo[txInfo.PairIndex].LpAmount, lpDeltaForFromAccount)

	// pool info
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

	finalPoolA := ffmath.Add(liquidityInfo.AssetA, txInfo.AssetAAmount)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, txInfo.AssetBAmount)
	poolDeltaForToAccount := &types.LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               txInfo.AssetAAmount,
		AssetBId:             txInfo.AssetBId,
		AssetB:               txInfo.AssetAAmount,
		LpAmount:             lpDeltaForFromAccount,
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

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       types.LiquidityAssetType,
		AccountIndex:    types.NilTxAccountIndex,
		AccountName:     types.NilAccountName,
		Balance:         basePool.String(),
		BalanceDelta:    poolDeltaForToAccount.String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: 0,
	})

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

func (e *AddLiquidityExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := legendTxTypes.ComputeAddLiquidityMsgHash(e.txInfo, mimc.NewMiMC())
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
		TxAmount:      e.txInfo.LpAmount.String(),
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
