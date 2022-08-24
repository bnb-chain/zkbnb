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
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

type WithdrawNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.WithdrawNftTxInfo
}

func NewWithdrawNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseWithdrawNftTxInfo(tx.TxInfo)
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
	err := e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}

	nftInfo := e.bc.StateDB().NftMap[txInfo.NftIndex]
	creatorAccount := e.bc.StateDB().AccountMap[txInfo.CreatorAccountIndex]

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

	fromAccount := e.bc.StateDB().AccountMap[txInfo.AccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	nftInfo := e.bc.StateDB().NftMap[txInfo.NftIndex]
	if nftInfo.OwnerAccountIndex != txInfo.AccountIndex {
		return errors.New("account is not owner of the nft")
	}

	return nil
}

func (e *WithdrawNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	oldNft := bc.StateDB().NftMap[txInfo.NftIndex]
	fromAccount := bc.StateDB().AccountMap[txInfo.AccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]

	// apply changes
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	newNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	bc.StateDB().NftMap[txInfo.NftIndex] = &nft.L2Nft{
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

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateNftIndexMap[txInfo.NftIndex] = statedb.StateCachePending
	stateCache.PriorityOperations++
	stateCache.PendingNewNftWithdrawHistory[txInfo.NftIndex] = &nft.L2NftWithdrawHistory{
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
	buf.WriteByte(uint8(types.TxTypeWithdrawNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk3 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(chunk3)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(common2.Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *WithdrawNftExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.AccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}

	err := e.bc.StateDB().UpdateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.StateDB().UpdateNftTree(txInfo.NftIndex)
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
	nftModel := e.bc.StateDB().NftMap[txInfo.NftIndex]

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
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
			types.ZeroBigInt,
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
		AssetType:    types.NftAssetType,
		AccountIndex: types.NilTxAccountIndex,
		AccountName:  types.NilAccountName,
		Balance: types.ConstructNftInfo(
			nftModel.NftIndex,
			nftModel.CreatorAccountIndex,
			nftModel.OwnerAccountIndex,
			nftModel.NftContentHash,
			nftModel.NftL1TokenId,
			nftModel.NftL1Address,
			nftModel.CreatorTreasuryRate,
			nftModel.CollectionId,
		).String(),
		BalanceDelta:    types.EmptyNftInfo(txInfo.NftIndex).String(),
		Order:           order,
		Nonce:           0,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: 0,
	})

	// create account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  creatorAccount.AccountName,
		Balance:      creatorAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			types.ZeroBigInt,
			types.ZeroBigInt,
			types.ZeroBigInt,
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
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails, nil
}

func (e *WithdrawNftExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := legendTxTypes.ComputeWithdrawNftMsgHash(e.txInfo, mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      e.txInfo.NftIndex,
		PairIndex:     types.NilPairIndex,
		AssetId:       types.NilAssetId,
		TxAmount:      types.NilAssetAmountStr,
		Memo:          "",
		NativeAddress: e.txInfo.ToAddress,
		AccountIndex:  e.txInfo.AccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
