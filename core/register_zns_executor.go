package core

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/util"
)

type RegisterZnsExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.RegisterZnsTxInfo
}

func NewRegisterZnsExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse register tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &RegisterZnsExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *RegisterZnsExecutor) Prepare() error {
	return nil
}

func (e *RegisterZnsExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	_, err := bc.AccountModel.GetAccountByName(txInfo.AccountName)
	if err != sqlx.ErrNotFound {
		return errors.New("invalid account name, already registered")
	}

	for index := range bc.stateCache.pendingNewAccountIndexMap {
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

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *RegisterZnsExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}

func (e *RegisterZnsExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	return nil, nil
}
