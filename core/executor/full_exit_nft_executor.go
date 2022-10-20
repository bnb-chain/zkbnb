package executor

import (
	"bytes"
	"encoding/json"
	"errors"
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

	txInfo *txtypes.FullExitNftTxInfo

	exitNft   *nft.L2Nft
	exitEmpty bool
}

func NewFullExitNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseFullExitNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse full exit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &FullExitNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *FullExitNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.StateDB().GetAccountByNameHash(accountNameHash)
	if err != nil {
		return err
	}

	// Set the right account index.
	txInfo.AccountIndex = account.AccountIndex

	// Default withdraw an empty nft.
	// Case1: the nft index isn't exist.
	// Case2: the account isn't the owner of the nft.
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	exitNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		NftL1Address:        emptyNftInfo.NftL1Address,
		NftL1TokenId:        emptyNftInfo.NftL1TokenId,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
	}

	var isExitEmptyNft = true
	nft, err := e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil && err != types.DbErrNotFound {
		return err
	}

	if err == nil && nft.OwnerAccountIndex == account.AccountIndex {
		// Set the right nft if the owner is correct.
		exitNft = nft
		isExitEmptyNft = false
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{types.EmptyAccountAssetId}) // Prepare asset 0 for generate an empty tx detail.
	err = e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Set the right tx info.
	txInfo.CreatorAccountIndex = exitNft.CreatorAccountIndex
	txInfo.CreatorTreasuryRate = exitNft.CreatorTreasuryRate
	txInfo.CreatorAccountNameHash = common.FromHex(types.EmptyAccountNameHash)
	if !isExitEmptyNft {
		creator, err := bc.StateDB().GetFormatAccount(exitNft.CreatorAccountIndex)
		if err != nil {
			return err
		}
		txInfo.CreatorAccountNameHash = common.FromHex(creator.AccountNameHash)
	}
	txInfo.NftL1Address = exitNft.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(exitNft.NftL1TokenId, 10)
	txInfo.NftContentHash = common.FromHex(exitNft.NftContentHash)
	txInfo.CollectionId = exitNft.CollectionId

	e.exitNft = exitNft
	e.exitEmpty = isExitEmptyNft
	return nil
}

func (e *FullExitNftExecutor) VerifyInputs(skipGasAmtChk bool) error {
	return nil
}

func (e *FullExitNftExecutor) ApplyTransaction() error {
	if e.exitEmpty {
		return nil
	}

	// Set nft to empty nft.
	txInfo := e.txInfo
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	emptyNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		NftL1Address:        emptyNftInfo.NftL1Address,
		NftL1TokenId:        emptyNftInfo.NftL1TokenId,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
	}

	stateCache := e.bc.StateDB()
	stateCache.SetPendingNft(txInfo.NftIndex, emptyNft)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *FullExitNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeFullExitNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *FullExitNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.NftIndex = e.txInfo.NftIndex
	e.tx.AccountIndex = e.txInfo.AccountIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *FullExitNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	exitAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return nil, err
	}
	txDetails := make([]*tx.TxDetail, 0, 2)

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
		AccountName:     exitAccount.AccountName,
		Balance:         baseBalance.String(),
		BalanceDelta:    emptyDelta.String(),
		AccountOrder:    accountOrder,
		Order:           order,
		Nonce:           exitAccount.Nonce,
		CollectionNonce: exitAccount.CollectionNonce,
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
			oldNft.NftL1TokenId,
			oldNft.NftL1Address,
			oldNft.CreatorTreasuryRate,
			oldNft.CollectionId,
		)
		if txInfo.AccountIndex != oldNft.OwnerAccountIndex {
			newNft = baseNft
		}
	}

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     exitAccount.AccountName,
		Balance:         baseNft.String(),
		BalanceDelta:    newNft.String(),
		AccountOrder:    types.NilAccountOrder,
		Order:           order,
		Nonce:           exitAccount.Nonce,
		CollectionNonce: exitAccount.CollectionNonce,
	})

	return txDetails, nil
}
