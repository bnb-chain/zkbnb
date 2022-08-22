package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type FullExitNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.FullExitNftTxInfo

	exitNft *nft.L2Nft
}

func NewFullExitNftExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseFullExitNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse full exit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &FullExitNftExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *FullExitNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.AccountModel.GetAccountByNameHash(accountNameHash)
	if err != nil {
		for index := range bc.stateCache.pendingNewAccountIndexMap {
			if accountNameHash == bc.accountMap[index].AccountNameHash {
				account, err = commonAsset.FromFormatAccountInfo(bc.accountMap[index])
				break
			}
		}

		if err != nil {
			return errors.New("invalid account name hash")
		}
	}

	// Set the right account index.
	txInfo.AccountIndex = account.AccountIndex

	// Default withdraw an empty nft.
	// Case1: the nft index isn't exist.
	// Case2: the account isn't the owner of the nft.
	emptyNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
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
	err = e.bc.prepareNft(txInfo.NftIndex)
	if err == nil && bc.nftMap[txInfo.NftIndex].OwnerAccountIndex == account.AccountIndex {
		// Set the right nft if the owner is correct.
		exitNft = bc.nftMap[txInfo.NftIndex]
	}

	accounts := []int64{account.AccountIndex, exitNft.CreatorAccountIndex}
	assets := []int64{0} // Just used for generate an empty tx detail.
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	// Set the right tx info.
	txInfo.CreatorAccountIndex = exitNft.CreatorAccountIndex
	txInfo.CreatorTreasuryRate = exitNft.CreatorTreasuryRate
	txInfo.CreatorAccountNameHash = common.FromHex(bc.accountMap[exitNft.CreatorAccountIndex].AccountNameHash)
	txInfo.NftL1Address = exitNft.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(exitNft.NftL1TokenId, 10)
	txInfo.NftContentHash = common.FromHex(exitNft.NftContentHash)
	txInfo.CollectionId = exitNft.CollectionId

	e.exitNft = exitNft
	return nil
}

func (e *FullExitNftExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.nftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.nftMap[txInfo.NftIndex].OwnerAccountIndex {
		// The check is not fully enough, just avoid explicit error.
		if common.Bytes2Hex(txInfo.NftContentHash) != commonConstant.NilNftContentHash {
			return errors.New("invalid nft content hash")
		}
	} else {
		// The check is not fully enough, just avoid explicit error.
		if common.Bytes2Hex(txInfo.NftContentHash) != bc.nftMap[txInfo.NftIndex].NftContentHash {
			return errors.New("invalid nft content hash")
		}
	}

	return nil
}

func (e *FullExitNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.nftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.nftMap[txInfo.NftIndex].OwnerAccountIndex {
		// Do nothing.
		return nil
	}

	// Set nft to empty nft.
	emptyNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
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
	bc.nftMap[txInfo.NftIndex] = emptyNft

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateNftIndexMap[txInfo.NftIndex] = StateCachePending
	stateCache.pendingNewNftWithdrawHistory = append(stateCache.pendingNewNftWithdrawHistory, &nft.L2NftWithdrawHistory{
		NftIndex:            bc.nftMap[txInfo.NftIndex].NftIndex,
		CreatorAccountIndex: bc.nftMap[txInfo.NftIndex].CreatorAccountIndex,
		OwnerAccountIndex:   bc.nftMap[txInfo.NftIndex].OwnerAccountIndex,
		NftContentHash:      bc.nftMap[txInfo.NftIndex].NftContentHash,
		NftL1Address:        bc.nftMap[txInfo.NftIndex].NftL1Address,
		NftL1TokenId:        bc.nftMap[txInfo.NftIndex].NftL1TokenId,
		CreatorTreasuryRate: bc.nftMap[txInfo.NftIndex].CreatorTreasuryRate,
		CollectionId:        bc.nftMap[txInfo.NftIndex].CollectionId,
	})

	return nil
}

func (e *FullExitNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeFullExitNft))
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
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(util.Uint256ToBytes(txInfo.NftL1TokenId))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pendingOnChainOperationsPubData = append(stateCache.pendingOnChainOperationsPubData, pubData)
	stateCache.pendingOnChainOperationsHash = util.ConcatKeccakHash(stateCache.pendingOnChainOperationsHash, pubData)
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *FullExitNftExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.nftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.nftMap[txInfo.NftIndex].OwnerAccountIndex {
		// Do nothing when nft state doesn't change.
		return nil
	}

	return bc.updateNftTree(txInfo.NftIndex)
}

func (e *FullExitNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *FullExitNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	bc := e.bc
	txInfo := e.txInfo
	exitAccount := e.bc.accountMap[txInfo.AccountIndex]
	txDetails := make([]*tx.TxDetail, 0, 2)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := exitAccount.AssetInfo[0]
	emptyDelta := &commonAsset.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      0,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  exitAccount.AccountName,
		Balance:      baseBalance.String(),
		BalanceDelta: emptyDelta.String(),
		AccountOrder: accountOrder,
		Order:        order,
	})
	// nft info
	order++
	newNft := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	if bc.nftMap[txInfo.NftIndex] != nil && txInfo.AccountIndex != bc.nftMap[txInfo.NftIndex].OwnerAccountIndex {
		newNft = commonAsset.ConstructNftInfo(
			bc.nftMap[txInfo.NftIndex].NftIndex,
			bc.nftMap[txInfo.NftIndex].CreatorAccountIndex,
			bc.nftMap[txInfo.NftIndex].OwnerAccountIndex,
			bc.nftMap[txInfo.NftIndex].NftContentHash,
			bc.nftMap[txInfo.NftIndex].NftL1TokenId,
			bc.nftMap[txInfo.NftIndex].NftL1Address,
			bc.nftMap[txInfo.NftIndex].CreatorTreasuryRate,
			bc.nftMap[txInfo.NftIndex].CollectionId,
		)
	}

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  exitAccount.AccountName,
		//Balance:      baseNft.String(), // Ignore base balance.
		BalanceDelta: newNft.String(),
		AccountOrder: commonConstant.NilAccountOrder,
		Order:        order,
	})

	return txDetails, nil
}
