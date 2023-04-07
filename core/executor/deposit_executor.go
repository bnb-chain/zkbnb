package executor

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type DepositExecutor struct {
	BaseExecutor

	TxInfo          *txtypes.DepositTxInfo
	IsCreateAccount bool
}

func NewDepositExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseDepositTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &DepositExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewDepositExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &DepositExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.DepositTxInfo),
	}, nil
}

func (e *DepositExecutor) SetTxInfo(info *txtypes.DepositTxInfo) {
	e.TxInfo = info
}

func (e *DepositExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.TxInfo

	// The account index from txInfo isn't true, find account by l1Address.
	l1Address := txInfo.L1Address
	account, err := bc.StateDB().GetAccountByL1Address(l1Address)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	if err == types.AppErrAccountNotFound {
		if !e.isDesertExit {
			if e.tx.Rollback == false {
				nextAccountIndex := e.bc.StateDB().GetNextAccountIndex()
				txInfo.AccountIndex = nextAccountIndex
			} else {
				//for rollback
				txInfo.AccountIndex = e.tx.AccountIndex
			}
		}
		e.IsCreateAccount = true
	} else {
		// Set the right account index.
		txInfo.AccountIndex = account.AccountIndex
	}

	if e.IsCreateAccount {
		err := e.CreateEmptyAccount(txInfo.AccountIndex, l1Address, []int64{txInfo.AssetId})
		if err != nil {
			return err
		}
	}
	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.AssetId})
	return e.BaseExecutor.Prepare()
}

func (e *DepositExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo

	if txInfo.AssetAmount.Cmp(types.ZeroBigInt) < 0 {
		return types.AppErrInvalidAssetAmount
	}

	return nil
}

func (e *DepositExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo
	var depositAccount *types.AccountInfo
	var err error
	if e.IsCreateAccount {
		depositAccount = e.GetCreatingAccount()
	} else {
		depositAccount, err = bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
		if err != nil {
			return err
		}
	}
	depositAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(depositAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(depositAccount.AccountIndex, depositAccount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *DepositExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeDeposit))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *DepositExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.AssetId = e.TxInfo.AssetId
	e.tx.TxAmount = e.TxInfo.AssetAmount.String()
	e.tx.IsCreateAccount = e.IsCreateAccount
	if e.tx.ToAccountIndex != e.iTxInfo.GetToAccountIndex() || e.IsCreateAccount {
		e.tx.IsPartialUpdate = true
	}
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *DepositExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo
	var depositAccount *types.AccountInfo
	var err error
	if e.IsCreateAccount {
		depositAccount = e.GetEmptyAccount()
	} else {
		depositAccount, err = e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
		if err != nil {
			return nil, err
		}
	}

	baseBalance := depositAccount.AssetInfo[txInfo.AssetId]
	deltaBalance := &types.AccountAsset{
		AssetId:                  txInfo.AssetId,
		Balance:                  txInfo.AssetAmount,
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails := make([]*tx.TxDetail, 0, 2)
	order := int64(0)
	accountOrder := int64(0)

	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.AssetId,
		AssetType:       types.FungibleAssetType,
		AccountIndex:    txInfo.AccountIndex,
		L1Address:       depositAccount.L1Address,
		Balance:         baseBalance.String(),
		BalanceDelta:    deltaBalance.String(),
		Order:           order,
		AccountOrder:    accountOrder,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
		PublicKey:       depositAccount.PublicKey,
	})
	if e.IsCreateAccount {
		order++
		txDetails = append(txDetails, &tx.TxDetail{
			AssetId:         types.EmptyAccountAssetId,
			AssetType:       types.CreateAccountType,
			AccountIndex:    txInfo.AccountIndex,
			L1Address:       depositAccount.L1Address,
			Balance:         depositAccount.L1Address,
			BalanceDelta:    txInfo.L1Address,
			Order:           order,
			AccountOrder:    0,
			Nonce:           depositAccount.Nonce,
			CollectionNonce: depositAccount.CollectionNonce,
			PublicKey:       depositAccount.PublicKey,
		})
	}
	return txDetails, nil
}

func (e *DepositExecutor) Finalize() error {
	if e.IsCreateAccount {
		bc := e.bc
		txInfo := e.TxInfo
		if !e.isDesertExit {
			bc.StateDB().AccountAssetTrees.UpdateCache(txInfo.AccountIndex, bc.CurrentBlock().BlockHeight)
		}
		accountInfo := e.GetCreatingAccount()
		bc.StateDB().SetPendingAccountL1AddressMap(accountInfo.L1Address, accountInfo.AccountIndex)
	}
	err := e.BaseExecutor.Finalize()
	if err != nil {
		return err
	}
	return nil
}
