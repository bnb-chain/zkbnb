package executor

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

type RegisterZnsExecutor struct {
	BaseExecutor

	txInfo *txtypes.RegisterZnsTxInfo
}

func NewRegisterZnsExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseRegisterZnsTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse register tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &RegisterZnsExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *RegisterZnsExecutor) Prepare() error {
	err := e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(e.txInfo.AccountIndex, []int64{})
	return nil
}

func (e *RegisterZnsExecutor) VerifyInputs(skipGasAmtChk bool) error {
	bc := e.bc
	txInfo := e.txInfo

	_, err := bc.StateDB().GetAccountByName(txInfo.AccountName)
	if err == nil {
		return errors.New("invalid account name, already registered")
	}

	if txInfo.AccountIndex != bc.StateDB().GetNextAccountIndex() {
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
		Nonce:           types.EmptyNonce,
		CollectionNonce: types.EmptyCollectionNonce,
		AssetInfo:       types.EmptyAccountAssetInfo,
		AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
		Status:          account.AccountStatusConfirmed,
	}
	formatAccount, err := chain.ToFormatAccountInfo(newAccount)
	if err != nil {
		return err
	}

	bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.AccountIndex, bc.CurrentBlock().BlockHeight)

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.AccountIndex, formatAccount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *RegisterZnsExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeRegisterZns))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize(common2.AccountNameToBytes32(txInfo.AccountName)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	pk, err := common2.ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("unable to parse pub key: %s", err.Error())
		return err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(common2.PrefixPaddingBufToChunkSize(pk.A.X.Marshal()))
	buf.Write(common2.PrefixPaddingBufToChunkSize(pk.A.Y.Marshal()))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *RegisterZnsExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.AccountIndex = e.txInfo.AccountIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *RegisterZnsExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}
