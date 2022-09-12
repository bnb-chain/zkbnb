package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type CreateCollectionExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.CreateCollectionTxInfo
}

func NewCreateCollectionExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseCreateCollectionTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &CreateCollectionExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *CreateCollectionExecutor) Prepare() error {
	txInfo := e.txInfo

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	err := e.BaseExecutor.Prepare(context.Background())
	if err != nil {
		return err
	}

	// Set the right collection nonce to tx info.
	txInfo.CollectionId = e.bc.StateDB().AccountMap[txInfo.AccountIndex].CollectionNonce
	return nil
}

func (e *CreateCollectionExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := e.bc.StateDB().AccountMap[txInfo.AccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	return nil
}

func (e *CreateCollectionExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.StateDB().AccountMap[txInfo.AccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]

	// apply changes
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++
	fromAccount.CollectionNonce++

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.AccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	return e.BaseExecutor.ApplyTransaction()
}

func (e *CreateCollectionExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeCreateCollection))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *CreateCollectionExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.CollectionId = e.txInfo.CollectionId

	return e.BaseExecutor.GetExecutedTx()
}

func (e *CreateCollectionExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := e.txInfo.Hash(mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      types.NilNftIndex,
		PairIndex:     types.NilPairIndex,
		AssetId:       types.NilAssetId,
		TxAmount:      "",
		Memo:          "",
		AccountIndex:  e.txInfo.AccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
