package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas/common/util"

	"github.com/bnb-chain/zkbas-crypto/ffmath"

	"github.com/bnb-chain/zkbas/common/commonAsset"

	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/zeromicro/go-zero/core/logx"
)

type DepositExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *commonTx.DepositTxInfo
}

func NewDepositExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &DepositExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *DepositExecutor) Prepare() error {
	bc := e.bc
	txInfo, err := commonTx.ParseDepositTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

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

	accounts := []int64{txInfo.AccountIndex}
	assets := []int64{txInfo.AssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	e.txInfo = txInfo
	return nil
}

func (e *DepositExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	if txInfo.AssetAmount.Cmp(ZeroBigInt) < 0 {
		return errors.New("invalid asset amount")
	}

	return nil
}

func (e *DepositExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	e.tx.TxDetails = e.GenerateTxDetails()

	depositAccount := bc.accountMap[txInfo.AccountIndex]
	depositAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(depositAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = StateCachePending
	return nil
}

func (e *DepositExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeDeposit))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(util.Uint128ToBytes(txInfo.AssetAmount))
	chunk1 := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.priorityOperations++
	stateCache.pubDataOffset = append(stateCache.pubDataOffset, uint32(len(stateCache.pubData)))
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *DepositExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.AccountIndex}
	assets := []int64{txInfo.AssetId}
	return bc.updateAccountTree(accounts, assets)
}

func (e *DepositExecutor) GetExecutedTx() (*tx.Tx, error) {
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

func (e *DepositExecutor) GenerateTxDetails() []*tx.TxDetail {
	txInfo := e.txInfo
	depositAccount := e.bc.accountMap[txInfo.AccountIndex]
	baseBalance := depositAccount.AssetInfo[txInfo.AssetId]
	deltaBalance := &commonAsset.AccountAsset{
		AssetId:                  txInfo.AssetId,
		Balance:                  txInfo.AssetAmount,
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetail := &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  depositAccount.AccountName,
		Balance:      baseBalance.String(),
		BalanceDelta: deltaBalance.String(),
		Order:        0,
		AccountOrder: 0,
	}
	return []*tx.TxDetail{txDetail}
}
