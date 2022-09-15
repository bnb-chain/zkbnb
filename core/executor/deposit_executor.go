package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type DepositExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.DepositTxInfo
}

func NewDepositExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseDepositTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &DepositExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *DepositExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.DB().AccountModel.GetAccountByNameHash(accountNameHash)
	if err != nil {
		exist := false
		for index := range bc.StateDB().PendingNewAccountMap {
			tempAccount, err := bc.StateDB().GetAccount(index)
			if err != nil {
				continue
			}
			if accountNameHash == tempAccount.AccountNameHash {
				account = tempAccount
				exist = true
				break
			}
		}

		if !exist {
			return errors.New("invalid account name hash")
		}
	}

	// Set the right account index.
	txInfo.AccountIndex = account.AccountIndex

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.AssetId})
	return e.BaseExecutor.Prepare()
}

func (e *DepositExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	if txInfo.AssetAmount.Cmp(types.ZeroBigInt) < 0 {
		return errors.New("invalid asset amount")
	}

	return nil
}

func (e *DepositExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	depositAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	depositAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(depositAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)

	stateCache := e.bc.StateDB()
	stateCache.SetPendingUpdateAccount(depositAccount.AccountIndex, depositAccount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *DepositExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeDeposit))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *DepositExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.AssetId = e.txInfo.AssetId
	e.tx.TxAmount = e.txInfo.AssetAmount.String()
	e.tx.AccountIndex = e.txInfo.AccountIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *DepositExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	depositAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return nil, err
	}
	baseBalance := depositAccount.AssetInfo[txInfo.AssetId]
	deltaBalance := &types.AccountAsset{
		AssetId:                  txInfo.AssetId,
		Balance:                  txInfo.AssetAmount,
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetail := &tx.TxDetail{
		AssetId:         txInfo.AssetId,
		AssetType:       types.FungibleAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     depositAccount.AccountName,
		Balance:         baseBalance.String(),
		BalanceDelta:    deltaBalance.String(),
		Order:           0,
		AccountOrder:    0,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
	}
	return []*tx.TxDetail{txDetail}, nil
}
