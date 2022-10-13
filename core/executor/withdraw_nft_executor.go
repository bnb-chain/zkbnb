package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
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

	txInfo *txtypes.WithdrawNftTxInfo
}

func NewWithdrawNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseWithdrawNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &WithdrawNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *WithdrawNftExecutor) Prepare() error {
	txInfo := e.txInfo

	nftInfo, err := e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	if nftInfo.CreatorAccountIndex != types.NilAccountIndex {
		e.MarkAccountAssetsDirty(nftInfo.CreatorAccountIndex, []int64{})
	}
	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Set the right details to tx info.
	txInfo.CreatorAccountIndex = nftInfo.CreatorAccountIndex
	txInfo.CreatorAccountNameHash = common.FromHex(types.EmptyAccountNameHash)
	if nftInfo.CreatorAccountIndex != types.NilAccountIndex {
		creatorAccount, err := e.bc.StateDB().GetFormatAccount(nftInfo.CreatorAccountIndex)
		if err != nil {
			return err
		}
		txInfo.CreatorAccountNameHash = common.FromHex(creatorAccount.AccountNameHash)
	}
	txInfo.CreatorTreasuryRate = nftInfo.CreatorTreasuryRate
	txInfo.NftContentHash = common.FromHex(nftInfo.NftContentHash)
	txInfo.NftL1Address = nftInfo.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(nftInfo.NftL1TokenId, 10)
	txInfo.CollectionId = nftInfo.CollectionId

	return nil
}

func (e *WithdrawNftExecutor) VerifyInputs(skipGasAmtChk bool) error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	nftInfo, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}
	if nftInfo.OwnerAccountIndex != txInfo.AccountIndex {
		return errors.New("account is not owner of the nft")
	}

	return nil
}

func (e *WithdrawNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

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
		NftL1Address:        newNftInfo.NftL1Address,
		NftL1TokenId:        newNftInfo.NftL1TokenId,
		CreatorTreasuryRate: newNftInfo.CreatorTreasuryRate,
		CollectionId:        newNftInfo.CollectionId,
	})
	stateCache.SetPendingUpdateGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
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

func (e *WithdrawNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.txInfo.GasFeeAssetId
	e.tx.GasFee = e.txInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.txInfo.NftIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *WithdrawNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
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
		AccountName:  fromAccount.AccountName,
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
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// nft delta
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    types.NftAssetType,
		AccountIndex: types.NilAccountIndex,
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
		AccountOrder:    types.NilAccountOrder,
		Nonce:           fromAccount.Nonce,
		CollectionNonce: fromAccount.CollectionNonce,
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
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           creatorAccount.Nonce,
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
		).String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           gasAccount.Nonce,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
	})
	return txDetails, nil
}
