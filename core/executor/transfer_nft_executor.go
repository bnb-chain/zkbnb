package executor

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type TransferNftExecutor struct {
	BaseExecutor

	TxInfo          *txtypes.TransferNftTxInfo
	IsCreateAccount bool
}

func NewTransferNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseTransferNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &TransferNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewTransferNftExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &TransferNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.TransferNftTxInfo),
	}, nil
}

func (e *TransferNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.TxInfo
	if !e.isDesertExit {
		txInfo.ToAccountIndex = types.NilAccountIndex
	}
	_, err := e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return err
	}

	toL1Address := txInfo.ToL1Address
	toAccount, err := bc.StateDB().GetAccountByL1Address(toL1Address)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	if err == types.AppErrAccountNotFound {
		if !e.isDesertExit {
			if !e.bc.StateDB().IsFromApi {
				if e.tx.Rollback == false {
					nextAccountIndex := e.bc.StateDB().GetNextAccountIndex()
					txInfo.ToAccountIndex = nextAccountIndex
				} else {
					//for rollback
					txInfo.ToAccountIndex = e.tx.ToAccountIndex
				}
			}
		}
		e.IsCreateAccount = true
	} else {
		txInfo.ToAccountIndex = toAccount.AccountIndex
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.FromAccountIndex, []int64{txInfo.GasFeeAssetId})
	// For empty tx details generation
	if e.IsCreateAccount {
		err := e.CreateEmptyAccount(txInfo.ToAccountIndex, txInfo.ToL1Address, []int64{types.EmptyAccountAssetId})
		if err != nil {
			return err
		}
	}
	e.MarkAccountAssetsDirty(txInfo.ToAccountIndex, []int64{types.EmptyAccountAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	return e.BaseExecutor.Prepare()
}

func (e *TransferNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo

	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}

	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}

	if !e.IsCreateAccount {
		toAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
		if err != nil {
			return err
		}
		if fromAccount.AccountIndex == toAccount.AccountIndex {
			return types.AppErrAccountInvalidToAccount
		}
		if txInfo.ToL1Address != toAccount.L1Address {
			return types.AppErrInvalidToAddress
		}
	}

	nft, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}
	if nft.OwnerAccountIndex != txInfo.FromAccountIndex {
		return errors.New("account is not owner of the nft")
	}

	return nil
}

func (e *TransferNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo
	var toAccount *types.AccountInfo

	fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	if err != nil {
		return err
	}
	nft, err := bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}

	if e.IsCreateAccount {
		toAccount = e.GetCreatingAccount()
	} else {
		toAccount, err = bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
		if err != nil {
			return err
		}
	}

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++
	nft.OwnerAccountIndex = txInfo.ToAccountIndex

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.FromAccountIndex, fromAccount)
	stateCache.SetPendingAccount(txInfo.ToAccountIndex, toAccount)
	stateCache.SetPendingNft(txInfo.NftIndex, nft)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *TransferNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeTransferNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *TransferNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.TxInfo.NftIndex
	e.tx.IsCreateAccount = e.IsCreateAccount
	if e.tx.ToAccountIndex != e.iTxInfo.GetToAccountIndex() || e.IsCreateAccount {
		e.tx.IsPartialUpdate = true
	}
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *TransferNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo
	nftModel, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return nil, err
	}
	var copiedAccounts map[int64]*types.AccountInfo
	var toAccount *types.AccountInfo
	if e.IsCreateAccount {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = e.GetEmptyAccount()
	} else {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex, txInfo.ToAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = copiedAccounts[txInfo.ToAccountIndex]
	}

	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 5)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      types.EmptyAccountAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		L1Address:    toAccount.L1Address,
		Balance:      toAccount.AssetInfo[types.EmptyAccountAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			types.EmptyAccountAssetId,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: toAccount.CollectionNonce,
		PublicKey:       toAccount.PublicKey,
	})
	if e.IsCreateAccount {
		order++
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:         types.EmptyAccountAssetId,
			AssetType:       types.CreateAccountType,
			AccountIndex:    txInfo.ToAccountIndex,
			L1Address:       toAccount.L1Address,
			Balance:         toAccount.L1Address,
			BalanceDelta:    txInfo.ToL1Address,
			Order:           order,
			AccountOrder:    accountOrder,
			Nonce:           toAccount.Nonce,
			CollectionNonce: toAccount.CollectionNonce,
			PublicKey:       toAccount.PublicKey,
		})
	}
	// to account nft delta
	oldNftInfo := &types.NftInfo{
		NftIndex:            nftModel.NftIndex,
		CreatorAccountIndex: nftModel.CreatorAccountIndex,
		OwnerAccountIndex:   nftModel.OwnerAccountIndex,
		NftContentHash:      nftModel.NftContentHash,
		RoyaltyRate:         nftModel.RoyaltyRate,
		CollectionId:        nftModel.CollectionId,
		NftContentType:      nftModel.NftContentType,
	}
	newNftInfo := &types.NftInfo{
		NftIndex:            nftModel.NftIndex,
		CreatorAccountIndex: nftModel.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      nftModel.NftContentHash,
		RoyaltyRate:         nftModel.RoyaltyRate,
		CollectionId:        nftModel.CollectionId,
		NftContentType:      nftModel.NftContentType,
	}
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.ToAccountIndex,
		L1Address:       toAccount.L1Address,
		Balance:         oldNftInfo.String(),
		BalanceDelta:    newNftInfo.String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: toAccount.CollectionNonce,
		PublicKey:       toAccount.PublicKey,
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
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
		PublicKey:       gasAccount.PublicKey,
	})
	return txDetails, nil
}

func (e *TransferNftExecutor) Finalize() error {
	if e.IsCreateAccount {
		bc := e.bc
		txInfo := e.TxInfo
		if !e.isDesertExit {
			bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.ToAccountIndex, bc.CurrentBlock().BlockHeight)
		}
		accountInfo := e.GetCreatingAccount()
		bc.StateDB().SetPendingAccountL1AddressMap(accountInfo.L1Address, accountInfo.AccountIndex)
	}
	err := e.BaseExecutor.Finalize()
	if err != nil {
		return err
	}
	return nil
}
