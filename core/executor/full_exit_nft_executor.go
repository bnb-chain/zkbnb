package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

type FullExitNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.FullExitNftTxInfo

	exitNft *nft.L2Nft
}

func NewFullExitNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseFullExitNftTxInfo(tx.TxInfo)
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
	account, err := bc.DB().AccountModel.GetAccountByNameHash(accountNameHash)
	if err != nil {
		for index := range bc.StateDB().PendingNewAccountIndexMap {
			if accountNameHash == bc.StateDB().AccountMap[index].AccountNameHash {
				account, err = chain.FromFormatAccountInfo(bc.StateDB().AccountMap[index])
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
	err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err == nil && bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex == account.AccountIndex {
		// Set the right nft if the owner is correct.
		exitNft = bc.StateDB().NftMap[txInfo.NftIndex]
	}

	accounts := []int64{account.AccountIndex, exitNft.CreatorAccountIndex}
	assets := []int64{0} // Just used for generate an empty tx detail.
	err = e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	// Set the right tx info.
	txInfo.CreatorAccountIndex = exitNft.CreatorAccountIndex
	txInfo.CreatorTreasuryRate = exitNft.CreatorTreasuryRate
	txInfo.CreatorAccountNameHash = common.FromHex(bc.StateDB().AccountMap[exitNft.CreatorAccountIndex].AccountNameHash)
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

	if bc.StateDB().NftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		// The check is not fully enough, just avoid explicit error.
		if common.Bytes2Hex(txInfo.NftContentHash) != types.NilNftContentHash {
			return errors.New("invalid nft content hash")
		}
	} else {
		// The check is not fully enough, just avoid explicit error.
		if common.Bytes2Hex(txInfo.NftContentHash) != bc.StateDB().NftMap[txInfo.NftIndex].NftContentHash {
			return errors.New("invalid nft content hash")
		}
	}

	return nil
}

func (e *FullExitNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.StateDB().NftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		// Do nothing.
		return nil
	}

	// Set nft to empty nft.
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
	bc.StateDB().NftMap[txInfo.NftIndex] = emptyNft

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateNftIndexMap[txInfo.NftIndex] = statedb.StateCachePending
	stateCache.PendingNewNftWithdrawHistory = append(stateCache.PendingNewNftWithdrawHistory, &nft.L2NftWithdrawHistory{
		NftIndex:            bc.StateDB().NftMap[txInfo.NftIndex].NftIndex,
		CreatorAccountIndex: bc.StateDB().NftMap[txInfo.NftIndex].CreatorAccountIndex,
		OwnerAccountIndex:   bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex,
		NftContentHash:      bc.StateDB().NftMap[txInfo.NftIndex].NftContentHash,
		NftL1Address:        bc.StateDB().NftMap[txInfo.NftIndex].NftL1Address,
		NftL1TokenId:        bc.StateDB().NftMap[txInfo.NftIndex].NftL1TokenId,
		CreatorTreasuryRate: bc.StateDB().NftMap[txInfo.NftIndex].CreatorTreasuryRate,
		CollectionId:        bc.StateDB().NftMap[txInfo.NftIndex].CollectionId,
	})

	return nil
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
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(common2.Uint256ToBytes(txInfo.NftL1TokenId))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *FullExitNftExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.StateDB().NftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		// Do nothing when nft state doesn't change.
		return nil
	}

	return bc.StateDB().UpdateNftTree(txInfo.NftIndex)
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
	exitAccount := e.bc.StateDB().AccountMap[txInfo.AccountIndex]
	txDetails := make([]*tx.TxDetail, 0, 2)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := exitAccount.AssetInfo[0]
	emptyDelta := &types.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      0,
		AssetType:    types.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  exitAccount.AccountName,
		Balance:      baseBalance.String(),
		BalanceDelta: emptyDelta.String(),
		AccountOrder: accountOrder,
		Order:        order,
	})
	// nft info
	order++
	newNft := types.EmptyNftInfo(txInfo.NftIndex)
	if bc.StateDB().NftMap[txInfo.NftIndex] != nil && txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		newNft = types.ConstructNftInfo(
			bc.StateDB().NftMap[txInfo.NftIndex].NftIndex,
			bc.StateDB().NftMap[txInfo.NftIndex].CreatorAccountIndex,
			bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex,
			bc.StateDB().NftMap[txInfo.NftIndex].NftContentHash,
			bc.StateDB().NftMap[txInfo.NftIndex].NftL1TokenId,
			bc.StateDB().NftMap[txInfo.NftIndex].NftL1Address,
			bc.StateDB().NftMap[txInfo.NftIndex].CreatorTreasuryRate,
			bc.StateDB().NftMap[txInfo.NftIndex].CollectionId,
		)
	}

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    types.NftAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  exitAccount.AccountName,
		//Balance:      baseNft.String(), // Ignore base balance.
		BalanceDelta: newNft.String(),
		AccountOrder: types.NilAccountOrder,
		Order:        order,
	})

	return txDetails, nil
}

func (e *FullExitNftExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	return nil, nil
}
