package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type AddLiquidityExecutor struct {
	bc          *BlockChain
	tx          *tx.Tx
	newPoolInfo *commonAsset.LiquidityInfo
	txInfo      *legendTxTypes.AddLiquidityTxInfo
}

func NewAddLiquidityExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &AddLiquidityExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *AddLiquidityExecutor) Prepare() error {
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	err = e.bc.prepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return err
	}

	liquidityModel := e.bc.liquidityMap[txInfo.PairIndex]

	accounts := []int64{txInfo.FromAccountIndex, liquidityModel.TreasuryAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex, txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	e.txInfo = txInfo
	return nil
}

func (e *AddLiquidityExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]
	if txInfo.ExpiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("expired tx")
	}

	if txInfo.Nonce != fromAccount.Nonce {
		return errors.New("invalid nonce")
	}

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

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return err
	}

	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		return errors.New("invalid liquidity")
	}

	err = txInfo.VerifySignature(fromAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (e *AddLiquidityExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// add details to tx info
	err := e.fillTxInfo()
	if err != nil {
		return err
	}

	// generate tx details
	e.tx.TxDetails = e.GenerateTxDetails()

	// apply changes
	fromAccountInfo := bc.accountMap[txInfo.FromAccountIndex]
	gasAccountInfo := bc.accountMap[txInfo.GasAccountIndex]
	liquidityModel := bc.liquidityMap[txInfo.PairIndex]
	treasuryAccount := bc.accountMap[liquidityModel.TreasuryAccountIndex]

	fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmount)
	treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount = ffmath.Add(treasuryAccount.AssetInfo[txInfo.PairIndex].LpAmount, txInfo.TreasuryAmount)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccountInfo.Nonce++

	bc.liquidityMap[txInfo.PairIndex] = &liquidity.Liquidity{
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

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[treasuryAccount.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingUpdateLiquidityIndexMap[txInfo.PairIndex] = StateCachePending
	return nil
}

func (e *AddLiquidityExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return err
	}

	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		txInfo.LpAmount, err = util.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("[ComputeEmptyLpAmount] : %v", err)
			return err
		}
	} else {
		txInfo.LpAmount, err = util.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			return err
		}
	}

	txInfo.AssetAId = liquidityInfo.AssetAId
	txInfo.AssetBId = liquidityInfo.AssetBId

	lpDeltaForTreasuryAccount, err := util.ComputeSLp(liquidityInfo.AssetA,
		liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		logx.Errorf("[ComputeSLp] err: %v", err)
		return err
	}

	// pool account pool info
	finalPoolA := ffmath.Add(liquidityInfo.AssetA, txInfo.AssetAAmount)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, txInfo.AssetBAmount)

	txInfo.TreasuryAmount = lpDeltaForTreasuryAccount
	txInfo.KLast, err = util.CleanPackedAmount(ffmath.Multiply(finalPoolA, finalPoolB))
	if err != nil {
		return err
	}

	return nil
}

func (e *AddLiquidityExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeAddLiquidity))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := util.AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(KLastBytes)
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	treasuryAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk2 := util.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *AddLiquidityExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]

	accounts := []int64{txInfo.FromAccountIndex, liquidityModel.TreasuryAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetAId, txInfo.AssetBId, txInfo.PairIndex, txInfo.GasFeeAssetId}

	err := bc.updateAccountTree(accounts, assets)
	if err != nil {
		return err
	}

	err = bc.updateLiquidityTree(txInfo.PairIndex)
	if err != nil {
		return err
	}

	return nil
}

func (e *AddLiquidityExecutor) GetExecutedTx() (*tx.Tx, error) {
	bc := e.bc
	txInfo := e.txInfo
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}
	stateRoot := bc.getStateRoot()
	e.tx.BlockHeight = bc.currentBlock.BlockHeight
	e.tx.StateRoot = stateRoot
	e.tx.TxInfo = string(txInfoBytes)
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *AddLiquidityExecutor) GenerateTxDetails() []*tx.TxDetail {
	txInfo := e.txInfo

	fromAccount := e.bc.accountMap[txInfo.FromAccountIndex]
	gasAccount := e.bc.accountMap[txInfo.GasAccountIndex]

	liquidityModel := e.bc.liquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		// todo return error
		return nil
	}
	treasuryAccount := e.bc.accountMap[liquidityInfo.TreasuryAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetAId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetAId,
			ffmath.Neg(txInfo.AssetAAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})

	// from account asset B
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetBId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetBId,
			ffmath.Neg(txInfo.AssetBAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})

	// from account asset gas
	baseBalance := fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance
	if txInfo.GasFeeAssetId == txInfo.AssetAId {
		baseBalance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.AssetAAmount)
	} else if txInfo.GasFeeAssetId == txInfo.AssetBId {
		baseBalance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.AssetBAmount)
	}
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			baseBalance,
			fromAccount.AssetInfo[txInfo.GasFeeAssetId].LpAmount,
			fromAccount.AssetInfo[txInfo.GasFeeAssetId].OfferCanceledOrFinalized,
		).String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})

	// from account lp
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, txInfo.TreasuryAmount)
	var lpDeltaForFromAccount *big.Int
	if liquidityInfo.AssetA.Cmp(ZeroBigInt) == 0 {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
		if err != nil {
			logx.Errorf("unable to compute lp delta: %s", err.Error())
			// todo return error
			return nil
		}
	} else {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(ffmath.Div(ffmath.Multiply(txInfo.AssetAAmount, poolLp), liquidityInfo.AssetA))
		if err != nil {
			logx.Errorf(" unable to compute lp delta: %s", err.Error())
			// todo return error
			return nil
		}
	}

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.AssetBId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex,
			ZeroBigInt,
			lpDeltaForFromAccount,
			ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
	})

	// pool info
	basePool, err := commonAsset.ConstructLiquidityInfo(
		e.bc.liquidityMap[txInfo.PairIndex].PairIndex,
		e.bc.liquidityMap[txInfo.PairIndex].AssetAId,
		e.bc.liquidityMap[txInfo.PairIndex].AssetA,
		e.bc.liquidityMap[txInfo.PairIndex].AssetBId,
		e.bc.liquidityMap[txInfo.PairIndex].AssetB,
		e.bc.liquidityMap[txInfo.PairIndex].LpAmount,
		e.bc.liquidityMap[txInfo.PairIndex].KLast,
		e.bc.liquidityMap[txInfo.PairIndex].FeeRate,
		e.bc.liquidityMap[txInfo.PairIndex].TreasuryAccountIndex,
		e.bc.liquidityMap[txInfo.PairIndex].TreasuryRate,
	)
	if err != nil {
		// todo return error
		return nil
	}

	finalPoolA := ffmath.Add(liquidityInfo.AssetA, txInfo.AssetAAmount)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, txInfo.AssetBAmount)
	poolDeltaForToAccount := &commonAsset.LiquidityInfo{
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
	newPool, err := commonAsset.ComputeNewBalance(
		commonAsset.LiquidityAssetType, basePool.String(), poolDeltaForToAccount.String())
	if err != nil {
		// todo return error
		return nil
	}

	newPoolInfo, err := commonAsset.ParseLiquidityInfo(newPool)
	if err != nil {
		// todo return error
		return nil
	}
	e.newPoolInfo = newPoolInfo

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       commonAsset.LiquidityAssetType,
		AccountIndex:    commonConstant.NilAccountIndex,
		AccountName:     commonConstant.NilAccountName,
		Balance:         basePool.String(),
		BalanceDelta:    poolDeltaForToAccount.String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    commonConstant.NilAccountOrder,
		CollectionNonce: 0,
	})

	// treasury account
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: treasuryAccount.AccountIndex,
		AccountName:  treasuryAccount.AccountNameHash,
		Balance:      basePool.String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, txInfo.TreasuryAmount, ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           treasuryAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: treasuryAccount.CollectionNonce,
	})

	// gas account asset gas
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails
}
