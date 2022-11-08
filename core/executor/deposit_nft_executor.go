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

type DepositNftExecutor struct {
	BaseExecutor

	txInfo *txtypes.DepositNftTxInfo
}

func NewDepositNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseDepositNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &DepositNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *DepositNftExecutor) Prepare() error {
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

	_, err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{types.EmptyAccountAssetId}) // Prepare asset 0 for generate an empty tx detail.
	return e.BaseExecutor.Prepare()
}

func (e *DepositNftExecutor) VerifyInputs(skipGasAmtChk bool) error {
	bc := e.bc
	txInfo := e.txInfo

	nft, err := bc.StateDB().GetNft(txInfo.NftIndex)
	if err != nil {
		return err
	}
	if nft.NftContentHash != types.EmptyNftContentHash {
		return types.AppErrNftAlreadyExist
	}

	return nil
}

func (e *DepositNftExecutor) ApplyTransaction() error {
	txInfo := e.txInfo

	nft := &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.AccountIndex,
		NftContentHash:      common.Bytes2Hex(txInfo.NftContentHash),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}

	stateCache := e.bc.StateDB()
	stateCache.SetPendingNft(txInfo.NftIndex, nft)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *DepositNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeDepositNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *DepositNftExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *DepositNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	depositAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return nil, err
	}
	txDetails := make([]*tx.TxDetail, 0, 2)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := depositAccount.AssetInfo[types.EmptyAccountAssetId]
	deltaBalance := &types.AccountAsset{
		AssetId:                  types.EmptyAccountAssetId,
		Balance:                  big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         types.EmptyAccountAssetId,
		AssetType:       types.FungibleAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     depositAccount.AccountName,
		Balance:         baseBalance.String(),
		BalanceDelta:    deltaBalance.String(),
		AccountOrder:    accountOrder,
		Order:           order,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
	})
	// nft info
	order++
	baseNft := types.EmptyNftInfo(txInfo.NftIndex)
	newNft := types.ConstructNftInfo(
		txInfo.NftIndex,
		txInfo.CreatorAccountIndex,
		txInfo.AccountIndex,
		common.Bytes2Hex(txInfo.NftContentHash),
		txInfo.CreatorTreasuryRate,
		txInfo.CollectionId,
	)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     depositAccount.AccountName,
		Balance:         baseNft.String(),
		BalanceDelta:    newNft.String(),
		AccountOrder:    types.NilAccountOrder,
		Order:           order,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
	})

	return txDetails, nil
}
