package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"

	"github.com/bnb-chain/zkbas/common/commonConstant"

	"github.com/bnb-chain/zkbas/common/model/liquidity"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type SwapExecutor struct {
	bc          *BlockChain
	tx          *tx.Tx
	newPoolInfo *commonAsset.LiquidityInfo
	txInfo      *legendTxTypes.SwapTxInfo
}

func NewSwapExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &SwapExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *SwapExecutor) Prepare() error {
	txInfo, err := commonTx.ParseSwapTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	accounts := []int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetAId, txInfo.AssetBId, txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	err = e.bc.prepareLiquidity(txInfo.PairIndex)
	if err != nil {
		logx.Errorf("prepare liquidity failed: %s", err.Error())
		return errors.New("internal error")
	}

	// check the other restrictions
	err = e.fillTxInfo()
	if err != nil {
		return err
	}

	return nil
}

func (e *SwapExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	if err := e.bc.verifyExpiredAt(txInfo.ExpiredAt); err != nil {
		return err
	}

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]

	if err := e.bc.verifyNonce(fromAccount.AccountIndex, txInfo.Nonce); err != nil {
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

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return err
	}

	if !((liquidityModel.AssetAId == txInfo.AssetAId && liquidityModel.AssetBId == txInfo.AssetBId) ||
		(liquidityModel.AssetAId == txInfo.AssetBId && liquidityModel.AssetBId == txInfo.AssetAId)) {
		return errors.New("invalid asset ids")
	}

	if liquidityInfo.AssetA == nil || liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil || liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 {
		return errors.New("liquidity is empty")
	}

	err = txInfo.VerifySignature(fromAccount.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func constructLiquidityInfo(liquidity *liquidity.Liquidity) (*commonAsset.LiquidityInfo, error) {
	return commonAsset.ConstructLiquidityInfo(
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
	fromAccountInfo := bc.accountMap[txInfo.FromAccountIndex]
	gasAccountInfo := bc.accountMap[txInfo.GasAccountIndex]

	fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance = ffmath.Add(fromAccountInfo.AssetInfo[txInfo.AssetBId].Balance, txInfo.AssetBAmountDelta)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccountInfo.Nonce++

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]
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
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingUpdateLiquidityIndexMap[txInfo.PairIndex] = StateCachePending
	return nil
}

func (e *SwapExecutor) fillTxInfo() error {
	bc := e.bc
	txInfo := e.txInfo

	liquidityModel := bc.liquidityMap[txInfo.PairIndex]

	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return err
	}

	// add details to tx info
	var toDelta *big.Int
	if liquidityInfo.AssetAId == txInfo.AssetAId && liquidityInfo.AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
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
		toDelta, _, err = util.ComputeDelta(
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
	buf.WriteByte(uint8(commonTx.TxTypeSwap))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountDeltaBytes, err := util.AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *SwapExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	accounts := []int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.AssetAId, txInfo.AssetBId, txInfo.GasFeeAssetId}

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

func (e *SwapExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *SwapExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo

	copiedAccounts, err := e.bc.deepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}

	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]
	liquidityModel := e.bc.liquidityMap[txInfo.PairIndex]
	liquidityInfo, err := constructLiquidityInfo(liquidityModel)
	if err != nil {
		logx.Errorf("construct liquidity info error, err: ", err.Error())
		return nil, err
	}

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
	fromAccount.AssetInfo[txInfo.AssetAId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetAId].Balance, txInfo.AssetAAmount)
	if fromAccount.AssetInfo[txInfo.AssetAId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient asset a balance")
	}

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
			txInfo.AssetBAmountDelta,
			ZeroBigInt,
			ZeroBigInt,
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
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
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
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// pool info
	var poolDelta *commonAsset.LiquidityInfo
	poolAssetBDelta := ffmath.Neg(txInfo.AssetBAmountDelta)
	if txInfo.AssetAId == liquidityInfo.AssetAId {
		poolDelta = &commonAsset.LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetAId,
			AssetA:               txInfo.AssetAAmount,
			AssetBId:             txInfo.AssetBId,
			AssetB:               poolAssetBDelta,
			LpAmount:             txVerification.ZeroBigInt,
			KLast:                txVerification.ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	} else if txInfo.AssetAId == liquidityInfo.AssetBId {
		poolDelta = &commonAsset.LiquidityInfo{
			PairIndex:            txInfo.PairIndex,
			AssetAId:             txInfo.AssetBId,
			AssetA:               poolAssetBDelta,
			AssetBId:             txInfo.AssetAId,
			AssetB:               txInfo.AssetAAmount,
			LpAmount:             txVerification.ZeroBigInt,
			KLast:                txVerification.ZeroBigInt,
			FeeRate:              liquidityInfo.FeeRate,
			TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
			TreasuryRate:         liquidityInfo.TreasuryRate,
		}
	}

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
		logx.Errorf("construct liquidity error, err: %s", err.Error())
		return nil, err
	}

	newPool, err := commonAsset.ComputeNewBalance(
		commonAsset.LiquidityAssetType, basePool.String(), poolDelta.String())
	if err != nil {
		return nil, err
	}

	nPoolInfo, err := commonAsset.ParseLiquidityInfo(newPool)
	if err != nil {
		return nil, err
	}
	e.newPoolInfo = nPoolInfo

	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.PairIndex,
		AssetType:       commonAsset.LiquidityAssetType,
		AccountIndex:    commonConstant.NilAccountIndex,
		AccountName:     commonConstant.NilAccountName,
		Balance:         basePool.String(),
		BalanceDelta:    poolDelta.String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    commonConstant.NilAccountOrder,
		CollectionNonce: 0,
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
	return txDetails, nil
}
