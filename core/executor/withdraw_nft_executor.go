package executor

import (
	"bytes"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type WithdrawNftExecutor struct {
	BaseExecutor

	TxInfo *txtypes.WithdrawNftTxInfo
}

func NewWithdrawNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseWithdrawNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &WithdrawNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewWithdrawNftExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &WithdrawNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.WithdrawNftTxInfo),
	}, nil
}

func (e *WithdrawNftExecutor) PreLoadAccountAndNft(accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
	txInfo := e.TxInfo
	accountIndexMap[txInfo.AccountIndex] = true
	accountIndexMap[txInfo.GasAccountIndex] = true
	nftIndexMap[txInfo.NftIndex] = true
}

func (e *WithdrawNftExecutor) Prepare() error {
	txInfo := e.TxInfo

	nftInfo, err := e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return types.AppErrPrepareNftFailed
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	if nftInfo.CreatorAccountIndex != types.NilAccountIndex {
		e.MarkAccountAssetsDirty(nftInfo.CreatorAccountIndex, []int64{types.EmptyAccountAssetId})
	}
	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Set the right details to tx info.
	txInfo.CreatorAccountIndex = nftInfo.CreatorAccountIndex
	txInfo.CreatorL1Address = types.EmptyL1Address
	if nftInfo.CreatorAccountIndex != types.NilAccountIndex {
		creatorAccount, err := e.bc.StateDB().GetFormatAccount(nftInfo.CreatorAccountIndex)
		if err != nil {
			return err
		}
		txInfo.CreatorL1Address = creatorAccount.L1Address
	}
	txInfo.RoyaltyRate = nftInfo.RoyaltyRate
	txInfo.NftContentHash = common.FromHex(nftInfo.NftContentHash)
	txInfo.CollectionId = nftInfo.CollectionId
	txInfo.NftContentType = nftInfo.NftContentType
	return nil
}

func (e *WithdrawNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}

	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}

	nftInfo, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}
	if nftInfo.OwnerAccountIndex != txInfo.AccountIndex {
		return types.AppErrNotNftOwner
	}

	return nil
}

func (e *WithdrawNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

	oldNft, err := bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}
	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}

	// apply changes
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	newNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.AccountIndex, fromAccount)
	stateCache.SetPendingNft(txInfo.NftIndex, &nft.L2Nft{
		Model:               oldNft.Model,
		NftIndex:            newNftInfo.NftIndex,
		CreatorAccountIndex: newNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   newNftInfo.OwnerAccountIndex,
		NftContentHash:      newNftInfo.NftContentHash,
		RoyaltyRate:         newNftInfo.RoyaltyRate,
		CollectionId:        newNftInfo.CollectionId,
		NftContentType:      newNftInfo.NftContentType,
	})
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *WithdrawNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeWithdrawNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(common2.AddressStrToBytes(txInfo.CreatorL1Address))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.WriteByte(uint8(txInfo.NftContentType))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *WithdrawNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.TxInfo.NftIndex
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *WithdrawNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo
	nftModel, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return nil, err
	}

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.CreatorAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}

	fromAccount := copiedAccounts[txInfo.AccountIndex]
	creatorAccount := copiedAccounts[txInfo.CreatorAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientGasFeeBalance
	}

	// nft delta
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    types.NftAssetType,
		AccountIndex: types.NilAccountIndex,
		L1Address:    types.NilL1Address,
		Balance: types.ConstructNftInfo(
			nftModel.NftIndex,
			nftModel.CreatorAccountIndex,
			nftModel.OwnerAccountIndex,
			nftModel.NftContentHash,
			nftModel.RoyaltyRate,
			nftModel.CollectionId,
			nftModel.NftContentType,
		).String(),
		BalanceDelta:    types.EmptyNftInfo(txInfo.NftIndex).String(),
		Order:           order,
		AccountOrder:    types.NilAccountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})

	// create account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      types.EmptyAccountAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		L1Address:    creatorAccount.L1Address,
		Balance:      creatorAccount.AssetInfo[types.EmptyAccountAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			types.EmptyAccountAssetId,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           creatorAccount.Nonce,
		CollectionNonce: creatorAccount.CollectionNonce,
		PublicKey:       creatorAccount.PublicKey,
	})

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		L1Address:    gasAccount.L1Address,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
		PublicKey:       gasAccount.PublicKey,
	})
	return txDetails, nil
}
