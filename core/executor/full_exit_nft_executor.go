package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/tree"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type FullExitNftExecutor struct {
	BaseExecutor

	TxInfo *txtypes.FullExitNftTxInfo

	exitNft         *nft.L2Nft
	exitEmpty       bool
	AccountNotExist bool
}

func NewFullExitNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseFullExitNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse full exit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &FullExitNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		TxInfo:       txInfo,
	}, nil
}

func (e *FullExitNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.TxInfo

	// The account index from txInfo isn't true, find account by l1Address.
	l1Address := txInfo.L1Address
	accountByL1Address, err := bc.StateDB().GetAccountByL1Address(l1Address)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	formatAccountByIndex, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	if formatAccountByIndex == nil {
		e.AccountNotExist = true
		return nil
	}

	var isExitEmptyNft = true

	// Default withdraw an empty nft.
	// Case1: the nft index isn't exist.
	// Case2: the account isn't the owner of the nft.
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	exitNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
		NftContentType:      emptyNftInfo.NftContentType,
	}

	if accountByL1Address != nil {
		if formatAccountByIndex.L1Address == accountByL1Address.L1Address &&
			formatAccountByIndex.AccountIndex == accountByL1Address.AccountIndex {
			nft, err := e.bc.StateDB().PrepareNft(txInfo.NftIndex)
			if err != nil && err != types.AppErrNftNotFound {
				return err
			}
			if err == nil && nft.OwnerAccountIndex == formatAccountByIndex.AccountIndex {
				// Set the right nft if the owner is correct.
				exitNft = nft
				isExitEmptyNft = false
			}
		}
	}

	// Mark the tree states that would be affected in this executor.
	if !isExitEmptyNft {
		e.MarkNftDirty(txInfo.NftIndex)
	}
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{types.EmptyAccountAssetId})        // Prepare asset 0 for generate an empty tx detail.
	e.MarkAccountAssetsDirty(txInfo.CreatorAccountIndex, []int64{types.EmptyAccountAssetId}) // Prepare asset 0 for generate an empty tx detail.

	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Set the right tx info.
	txInfo.CreatorAccountIndex = exitNft.CreatorAccountIndex
	txInfo.CreatorTreasuryRate = exitNft.CreatorTreasuryRate
	creator, err := bc.StateDB().GetFormatAccount(exitNft.CreatorAccountIndex)
	if err != nil {
		return err
	}
	txInfo.CreatorL1Address = creator.L1Address
	txInfo.NftContentHash = common.FromHex(exitNft.NftContentHash)
	txInfo.CollectionId = exitNft.CollectionId

	e.exitNft = exitNft
	e.exitEmpty = isExitEmptyNft
	return nil
}

func (e *FullExitNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	return nil
}

func (e *FullExitNftExecutor) ApplyTransaction() error {
	if e.exitEmpty || e.AccountNotExist {
		return nil
	}

	// Set nft to empty nft.
	txInfo := e.TxInfo
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	emptyNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
		NftContentType:      emptyNftInfo.NftContentType,
	}
	cacheNft, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err == nil {
		emptyNft.ID = cacheNft.ID
	}
	stateCache := e.bc.StateDB()
	stateCache.SetPendingNft(txInfo.NftIndex, emptyNft)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *FullExitNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeFullExitNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
	buf.Write(common2.AddressStrToBytes(txInfo.CreatorL1Address))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.WriteByte(uint8(txInfo.NftContentType))
	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *FullExitNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.NftIndex = e.TxInfo.NftIndex
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *FullExitNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.TxInfo
	var exitAccount *types.AccountInfo
	var err error
	if e.AccountNotExist {
		newAccount := chain.EmptyAccount(txInfo.AccountIndex, types.EmptyL1Address, tree.NilAccountAssetRoot)
		exitAccount, err = chain.ToFormatAccountInfo(newAccount)
		if err != nil {
			return nil, err
		}
	} else {
		exitAccount, err = e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
		if err != nil {
			return nil, err
		}
	}

	creatorAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.CreatorAccountIndex)
	if err != nil {
		return nil, err
	}

	txDetails := make([]*tx.TxDetail, 0, 3)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := exitAccount.AssetInfo[types.EmptyAccountAssetId]
	emptyDelta := &types.AccountAsset{
		AssetId:                  types.EmptyAccountAssetId,
		Balance:                  big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         types.EmptyAccountAssetId,
		AssetType:       types.FungibleAssetType,
		AccountIndex:    txInfo.AccountIndex,
		L1Address:       exitAccount.L1Address,
		Balance:         baseBalance.String(),
		BalanceDelta:    emptyDelta.String(),
		AccountOrder:    accountOrder,
		Order:           order,
		Nonce:           exitAccount.Nonce,
		CollectionNonce: exitAccount.CollectionNonce,
		PublicKey:       exitAccount.PublicKey,
	})
	// nft info
	order++
	emptyNft := types.EmptyNftInfo(txInfo.NftIndex)
	baseNft := emptyNft
	newNft := emptyNft
	oldNft, _ := bc.StateDB().GetNft(txInfo.NftIndex)
	if oldNft != nil {
		baseNft = types.ConstructNftInfo(
			oldNft.NftIndex,
			oldNft.CreatorAccountIndex,
			oldNft.OwnerAccountIndex,
			oldNft.NftContentHash,
			oldNft.CreatorTreasuryRate,
			oldNft.CollectionId,
			oldNft.NftContentType,
		)
		if txInfo.AccountIndex != oldNft.OwnerAccountIndex {
			newNft = baseNft
		}
	}

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.AccountIndex,
		L1Address:       exitAccount.L1Address,
		Balance:         baseNft.String(),
		BalanceDelta:    newNft.String(),
		AccountOrder:    types.NilAccountOrder,
		Order:           order,
		Nonce:           exitAccount.Nonce,
		CollectionNonce: exitAccount.CollectionNonce,
		PublicKey:       exitAccount.PublicKey,
	})

	// create account empty delta
	order++
	accountOrder++
	creatorAccountBalance := creatorAccount.AssetInfo[types.EmptyAccountAssetId]
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      types.EmptyAccountAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		L1Address:    creatorAccount.L1Address,
		Balance:      creatorAccountBalance.String(),
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

	return txDetails, nil
}
