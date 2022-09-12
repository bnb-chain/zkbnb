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

type WithdrawExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.WithdrawTxInfo
}

func NewWithdrawExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseWithdrawTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &WithdrawExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *WithdrawExecutor) Prepare() error {
	return e.BaseExecutor.Prepare(context.Background())
}

func (e *WithdrawExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := e.bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	if txInfo.GasFeeAssetId != txInfo.AssetId {
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			return errors.New("invalid asset amount")
		}
		if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
			return errors.New("invalid gas asset amount")
		}
	} else {
		deltaBalance := ffmath.Add(txInfo.AssetAmount, txInfo.GasFeeAssetAmount)
		if fromAccount.AssetInfo[txInfo.AssetId].Balance.Cmp(deltaBalance) < 0 {
			return errors.New("invalid asset amount")
		}
	}

	return nil
}

func (e *WithdrawExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.StateDB().AccountMap[txInfo.FromAccountIndex]
	gasAccount := bc.StateDB().AccountMap[txInfo.GasAccountIndex]

	// apply changes
	fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = statedb.StateCachePending
	stateCache.PendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = statedb.StateCachePending
	return e.BaseExecutor.ApplyTransaction()
}

func (e *WithdrawExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeWithdraw))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *WithdrawExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *WithdrawExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
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
		AssetId:       e.txInfo.AssetId,
		TxAmount:      e.txInfo.AssetAmount.String(),
		Memo:          "",
		NativeAddress: e.txInfo.ToAddress,
		AccountIndex:  e.txInfo.FromAccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: types.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
