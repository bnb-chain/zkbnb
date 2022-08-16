package core

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/zkbas/common/util"

	"github.com/bnb-chain/zkbas/common/commonAsset"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/ethereum/go-ethereum/common"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterZnsExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *commonTx.RegisterZnsTxInfo
}

func NewRegisterZnsExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &RegisterZnsExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *RegisterZnsExecutor) Prepare() error {
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse register tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	e.txInfo = txInfo
	return nil
}

func (e *RegisterZnsExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	_, err := bc.AccountModel.GetAccountByAccountName(txInfo.AccountName)
	if err != sqlx.ErrNotFound {
		return errors.New("invalid account name, already registered")
	}

	for index, _ := range bc.stateCache.pendingNewAccountIndexMap {
		if txInfo.AccountName == bc.accountMap[index].AccountName {
			return errors.New("invalid account name, already registered")
		}
	}

	if txInfo.AccountIndex != bc.getNextAccountIndex() {
		return errors.New("invalid account index")
	}

	return nil
}

func (e *RegisterZnsExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo
	var err error

	e.tx.TxDetails = e.GenerateTxDetails()

	newAccount := &account.Account{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     txInfo.AccountName,
		PublicKey:       txInfo.PubKey,
		AccountNameHash: common.Bytes2Hex(txInfo.AccountNameHash),
		L1Address:       e.tx.NativeAddress,
		Nonce:           commonConstant.NilNonce,
		CollectionNonce: commonConstant.NilNonce,
		AssetInfo:       commonConstant.NilAssetInfo,
		AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
		Status:          account.AccountStatusConfirmed,
	}
	bc.accountMap[txInfo.AccountIndex], err = commonAsset.ToFormatAccountInfo(newAccount)
	if err != nil {
		return err
	}

	stateCache := e.bc.stateCache
	stateCache.pendingNewAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	return nil
}

func (e *RegisterZnsExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeRegisterZns))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize(util.AccountNameToBytes32(txInfo.AccountName)))
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	pk, err := util.ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("unable to parse pub key: %s", err.Error())
		return err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(util.PrefixPaddingBufToChunkSize(pk.A.X.Marshal()))
	buf.Write(util.PrefixPaddingBufToChunkSize(pk.A.Y.Marshal()))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *RegisterZnsExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.AccountIndex}

	emptyAssetTree, err := tree.NewEmptyAccountAssetTree(bc.treeCtx, txInfo.AccountIndex, uint64(bc.currentBlock.BlockHeight))
	if err != nil {
		logx.Errorf("new empty account asset tree failed: %s", err.Error())
		return err
	}
	bc.accountAssetTrees = append(bc.accountAssetTrees, emptyAssetTree)

	return bc.updateAccountTree(accounts, nil)
}

func (e *RegisterZnsExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.BlockHeight = e.bc.currentBlock.BlockHeight
	e.tx.StateRoot = e.bc.getStateRoot()
	e.tx.TxInfo = string(txInfoBytes)
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *RegisterZnsExecutor) GenerateTxDetails() []*tx.TxDetail {
	return nil
}
