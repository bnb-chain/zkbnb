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

type MintNftExecutor struct {
	BaseExecutor

	TxInfo          *txtypes.MintNftTxInfo
	IsCreateAccount bool
}

func NewMintNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseMintNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &MintNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewMintNftExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &MintNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.MintNftTxInfo),
	}, nil
}

func (e *MintNftExecutor) Prepare() error {
	txInfo := e.TxInfo
	if !e.isDesertExit {
		txInfo.ToAccountIndex = types.NilAccountIndex
	}

	if !e.bc.StateDB().IsFromApi {
		if !e.isDesertExit {
			// Set the right nft index for tx info.
			if e.tx.Rollback == false {
				nextNftIndex := e.bc.StateDB().GetNextNftIndex()
				txInfo.NftIndex = nextNftIndex
			} else {
				//for rollback
				nextNftIndex := e.tx.NftIndex
				txInfo.NftIndex = nextNftIndex
			}
			logx.Infof("mint nft,pool id =%d,new nft=%d,BlockHeight=%d", e.tx.ID, txInfo.NftIndex, e.bc.CurrentBlock().BlockHeight)
		}
		// Mark the tree states that would be affected in this executor.
		e.MarkNftDirty(txInfo.NftIndex)
	}

	toL1Address := txInfo.ToL1Address
	toAccount, err := e.bc.StateDB().GetAccountByL1Address(toL1Address)
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

	e.MarkAccountAssetsDirty(txInfo.CreatorAccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	// For empty tx details generation
	if e.IsCreateAccount {
		err := e.CreateEmptyAccount(txInfo.ToAccountIndex, txInfo.ToL1Address, []int64{types.EmptyAccountAssetId})
		if err != nil {
			return err
		}
	}
	e.MarkAccountAssetsDirty(txInfo.ToAccountIndex, []int64{})
	return e.BaseExecutor.Prepare()
}

func (e *MintNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo
	if err := e.Validate(); err != nil {
		return err
	}
	//if txInfo.CreatorAccountIndex != txInfo.ToAccountIndex {
	//	return types.AppErrInvalidToAccount
	//}
	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	if txInfo.NftContentType != 0 && txInfo.NftContentType != 1 {
		return types.AppErrInvalidNftContentType
	}

	if !e.bc.StateDB().IsFromApi {
		nftContentHashLen := len(common.FromHex(txInfo.NftContentHash))
		if nftContentHashLen < 1 || nftContentHashLen > 32 {
			return types.AppErrInvalidNftContenthash
		}
	}

	creatorAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.CreatorAccountIndex)
	if err != nil {
		return err
	}
	if creatorAccount.CollectionNonce <= txInfo.NftCollectionId {
		return types.AppErrInvalidCollectionId
	}
	if creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}

	if !e.IsCreateAccount {
		toAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
		if err != nil {
			return err
		}
		if txInfo.ToL1Address != toAccount.L1Address {
			return types.AppErrInvalidToAddress
		}
	}

	return nil
}

func (e *MintNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo
	var toAccount *types.AccountInfo

	// apply changes
	creatorAccount, err := bc.StateDB().GetFormatAccount(txInfo.CreatorAccountIndex)
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

	creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	creatorAccount.Nonce++
	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.CreatorAccountIndex, creatorAccount)
	stateCache.SetPendingNft(txInfo.NftIndex, &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		RoyaltyRate:         txInfo.RoyaltyRate,
		CollectionId:        txInfo.NftCollectionId,
		NftContentType:      txInfo.NftContentType,
	})
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	stateCache.SetPendingAccount(txInfo.ToAccountIndex, toAccount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *MintNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeMintNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.NftCollectionId)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
	buf.WriteByte(uint8(txInfo.NftContentType))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *MintNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.TxInfo.NftIndex
	e.tx.IsPartialUpdate = true
	e.tx.IsCreateAccount = e.IsCreateAccount
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *MintNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

	var err error
	var copiedAccounts map[int64]*types.AccountInfo
	var toAccount *types.AccountInfo
	if e.IsCreateAccount {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.CreatorAccountIndex, txInfo.GasAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = e.GetEmptyAccount()
	} else {
		copiedAccounts, err = e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.CreatorAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex})
		if err != nil {
			return nil, err
		}
		toAccount = copiedAccounts[txInfo.ToAccountIndex]
	}

	creatorAccount := copiedAccounts[txInfo.CreatorAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 5)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		L1Address:    creatorAccount.L1Address,
		Balance:      creatorAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           creatorAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: creatorAccount.CollectionNonce,
		PublicKey:       creatorAccount.PublicKey,
	})
	creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientGasFeeBalance
	}

	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		L1Address:    toAccount.L1Address,
		Balance:      toAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
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
	oldNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	newNftInfo := &types.NftInfo{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		RoyaltyRate:         txInfo.RoyaltyRate,
		CollectionId:        txInfo.NftCollectionId,
		NftContentType:      txInfo.NftContentType,
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

func (e *MintNftExecutor) Validate() error {
	if len(e.TxInfo.MetaData) > 2000 {
		return types.AppErrInvalidMetaData.RefineError(2000)
	}
	if len(e.TxInfo.MutableAttributes) > 2000 {
		return types.AppErrInvalidMutableAttributes.RefineError(2000)
	}

	if e.IsCreateAccount {
		bc := e.bc
		txInfo := e.TxInfo
		if !e.isDesertExit {
			bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.ToAccountIndex, bc.CurrentBlock().BlockHeight)
			logx.Infof("create account,pool id =%d,new AccountIndex=%d,BlockHeight=%d", e.tx.ID, txInfo.ToAccountIndex, bc.CurrentBlock().BlockHeight)
		}
		accountInfo := e.GetCreatingAccount()
		bc.StateDB().SetPendingAccountL1AddressMap(accountInfo.L1Address, accountInfo.AccountIndex)
	}
	if !e.isDesertExit {
		if e.tx.Rollback == false {
			e.bc.StateDB().UpdateNftIndex(e.TxInfo.NftIndex)
		}
	}
	err := e.BaseExecutor.Finalize()
	if err != nil {
		return err
	}
	return nil
}
