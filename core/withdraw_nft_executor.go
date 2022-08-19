package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type WithdrawNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.WithdrawNftTxInfo
}

func NewWithdrawNftExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &WithdrawNftExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *WithdrawNftExecutor) Prepare() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.AccountIndex, txInfo.CreatorAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}
	err := e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	err = e.bc.prepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}

	nftInfo := e.bc.nftMap[txInfo.NftIndex]
	creatorAccount := e.bc.accountMap[txInfo.CreatorAccountIndex]

	// add details to tx info
	txInfo.CreatorAccountIndex = nftInfo.CreatorAccountIndex
	txInfo.CreatorAccountNameHash = common.FromHex(creatorAccount.AccountNameHash)
	txInfo.CreatorTreasuryRate = nftInfo.CreatorTreasuryRate
	txInfo.NftContentHash = common.FromHex(nftInfo.NftContentHash)
	txInfo.NftL1Address = nftInfo.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(nftInfo.NftL1TokenId, 10)
	txInfo.CollectionId = nftInfo.CollectionId

	return nil
}

func (e *WithdrawNftExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := e.bc.accountMap[txInfo.AccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	nftInfo := e.bc.nftMap[txInfo.NftIndex]
	if nftInfo.OwnerAccountIndex != txInfo.AccountIndex {
		return errors.New("account is not owner of the nft")
	}

	return nil
}

func (e *WithdrawNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	oldNft := bc.nftMap[txInfo.NftIndex]
	fromAccount := bc.accountMap[txInfo.AccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]

	// apply changes
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	newNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	bc.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		Model:               oldNft.Model,
		NftIndex:            newNftInfo.NftIndex,
		CreatorAccountIndex: newNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   newNftInfo.OwnerAccountIndex,
		NftContentHash:      newNftInfo.NftContentHash,
		NftL1Address:        newNftInfo.NftL1Address,
		NftL1TokenId:        newNftInfo.NftL1TokenId,
		CreatorTreasuryRate: newNftInfo.CreatorTreasuryRate,
		CollectionId:        newNftInfo.CollectionId,
	}

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingUpdateNftIndexMap[txInfo.NftIndex] = StateCachePending
	stateCache.priorityOperations++
	stateCache.pendingNewNftWithdrawHistory[txInfo.NftIndex] = &nft.L2NftWithdrawHistory{
		NftIndex:            oldNft.NftIndex,
		CreatorAccountIndex: oldNft.CreatorAccountIndex,
		OwnerAccountIndex:   oldNft.OwnerAccountIndex,
		NftContentHash:      oldNft.NftContentHash,
		NftL1Address:        oldNft.NftL1Address,
		NftL1TokenId:        oldNft.NftL1TokenId,
		CreatorTreasuryRate: oldNft.CreatorTreasuryRate,
		CollectionId:        oldNft.CollectionId,
	}

	return nil
}

func (e *WithdrawNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeWithdrawNft))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(util.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(util.AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := util.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(util.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk3 := util.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(chunk3)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(util.Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pendingOnChainOperationsPubData = append(stateCache.pendingOnChainOperationsPubData, pubData)
	stateCache.pendingOnChainOperationsHash = util.ConcatKeccakHash(stateCache.pendingOnChainOperationsHash, pubData)
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *WithdrawNftExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}

	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.updateNftTree(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("update nft tree error, err: %s", err.Error())
		return err
	}
	return nil
}

func (e *WithdrawNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *WithdrawNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	nftModel := e.bc.nftMap[txInfo.NftIndex]

	copiedAccounts, err := e.bc.deepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.CreatorAccountIndex, txInfo.GasAccountIndex})
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
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// nft delta
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		Balance: commonAsset.ConstructNftInfo(
			nftModel.NftIndex,
			nftModel.CreatorAccountIndex,
			nftModel.OwnerAccountIndex,
			nftModel.NftContentHash,
			nftModel.NftL1TokenId,
			nftModel.NftL1Address,
			nftModel.CreatorTreasuryRate,
			nftModel.CollectionId,
		).String(),
		BalanceDelta:    commonAsset.EmptyNftInfo(txInfo.NftIndex).String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    commonConstant.NilAccountOrder,
		CollectionNonce: 0,
	})

	// create account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  creatorAccount.AccountName,
		Balance:      creatorAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ZeroBigInt,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           creatorAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: creatorAccount.CollectionNonce,
	})

	// gas account gas asset
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
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails, nil
}
