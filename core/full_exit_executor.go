package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type FullExitExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.FullExitTxInfo
}

func NewFullExitExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseFullExitTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse full exit tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &FullExitExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *FullExitExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.AccountModel.GetAccountByNameHash(accountNameHash)
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

	accounts := []int64{txInfo.AccountIndex}
	assets := []int64{txInfo.AssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	// Set the right asset amount.
	txInfo.AssetAmount = bc.accountMap[txInfo.AccountIndex].AssetInfo[txInfo.AssetId].Balance

	return nil
}

func (e *FullExitExecutor) VerifyInputs() error {
	return nil
}

func (e *FullExitExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	exitAccount := bc.accountMap[txInfo.AccountIndex]
	exitAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(exitAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	return nil
}

func (e *FullExitExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeFullExit))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(util.Uint128ToBytes(txInfo.AssetAmount))
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pendingOnChainOperationsPubData = append(stateCache.pendingOnChainOperationsPubData, pubData)
	stateCache.pendingOnChainOperationsHash = util.ConcatKeccakHash(stateCache.pendingOnChainOperationsHash, pubData)
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *FullExitExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.AccountIndex}
	assets := []int64{txInfo.AssetId}
	return bc.updateAccountTree(accounts, assets)
}

func (e *FullExitExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *FullExitExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	exitAccount := e.bc.accountMap[txInfo.AccountIndex]
	baseBalance := exitAccount.AssetInfo[txInfo.AssetId]
	deltaBalance := &commonAsset.AccountAsset{
		AssetId:                  txInfo.AssetId,
		Balance:                  ffmath.Neg(txInfo.AssetAmount),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetail := &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  exitAccount.AccountName,
		Balance:      baseBalance.String(),
		BalanceDelta: deltaBalance.String(),
		Order:        0,
		AccountOrder: 0,
	}
	return []*tx.TxDetail{txDetail}, nil
}
