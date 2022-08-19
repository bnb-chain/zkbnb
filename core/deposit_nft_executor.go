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

type DepositNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.DepositNftTxInfo

	isNewNft bool
}

func NewDepositNftExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseDepositNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &DepositNftExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *DepositNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.AccountModel.GetAccountByAccountNameHash(accountNameHash)
	if err != nil {
		for index, _ := range bc.stateCache.pendingNewAccountIndexMap {
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

	accounts := []int64{txInfo.AccountIndex, txInfo.CreatorAccountIndex}
	assets := []int64{0} // Just used for generate an empty tx detail.
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	// Check if it is a new nft, or it is a nft previously withdraw from layer2.
	if txInfo.NftIndex == 0 && txInfo.CollectionId == 0 && txInfo.CreatorAccountIndex == 0 && txInfo.CreatorTreasuryRate == 0 {
		e.isNewNft = true
		// Set new nft index for new nft.
		txInfo.NftIndex = bc.getNextNftIndex()
	} else {
		err = e.bc.prepareNft(txInfo.NftIndex)
		if err != nil {
			logx.Errorf("prepare nft failed")
			return err
		}
	}

	return nil
}

func (e *DepositNftExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	if e.isNewNft {
		if bc.nftMap[txInfo.NftIndex] != nil {
			return errors.New("invalid nft index, already exist")
		}
	} else {
		if bc.nftMap[txInfo.NftIndex].OwnerAccountIndex != commonConstant.NilAccountIndex {
			return errors.New("invalid nft index, already exist")
		}
	}

	return nil
}

func (e *DepositNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	bc.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.AccountIndex,
		NftContentHash:      common.Bytes2Hex(txInfo.NftContentHash),
		NftL1Address:        txInfo.NftL1Address,
		NftL1TokenId:        txInfo.NftL1TokenId.String(),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}

	stateCache := e.bc.stateCache
	if e.isNewNft {
		stateCache.pendingNewNftIndexMap[txInfo.NftIndex] = StateCachePending
	} else {
		stateCache.pendingUpdateNftIndexMap[txInfo.NftIndex] = StateCachePending
	}

	return nil
}

func (e *DepositNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeDepositNft))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(util.AddressStrToBytes(txInfo.NftL1Address))
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(util.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk2 := util.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(util.Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *DepositNftExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	return bc.updateNftTree(txInfo.NftIndex)
}

func (e *DepositNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *DepositNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	depositAccount := e.bc.accountMap[txInfo.AccountIndex]
	txDetails := make([]*tx.TxDetail, 0, 2)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := depositAccount.AssetInfo[0]
	deltaBalance := &commonAsset.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      0,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  depositAccount.AccountName,
		Balance:      baseBalance.String(),
		BalanceDelta: deltaBalance.String(),
		AccountOrder: accountOrder,
		Order:        order,
	})
	// nft info
	order++
	baseNft := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	newNft := commonAsset.ConstructNftInfo(
		txInfo.NftIndex,
		txInfo.CreatorAccountIndex,
		txInfo.AccountIndex,
		common.Bytes2Hex(txInfo.NftContentHash),
		txInfo.NftL1TokenId.String(),
		txInfo.NftL1Address,
		txInfo.CreatorTreasuryRate,
		txInfo.CollectionId,
	)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  depositAccount.AccountName,
		Balance:      baseNft.String(),
		BalanceDelta: newNft.String(),
		AccountOrder: commonConstant.NilAccountOrder,
		Order:        order,
	})

	return txDetails, nil
}
