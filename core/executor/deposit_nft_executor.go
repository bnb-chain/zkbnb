package executor

import (
	"bytes"
	"encoding/json"
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

	TxInfo          *txtypes.DepositNftTxInfo
	IsCreateAccount bool
}

func NewDepositNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseDepositNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit nft tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &DepositNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func NewDepositNftExecutorForDesert(bc IBlockchain, txInfo txtypes.TxInfo) (TxExecutor, error) {
	return &DepositNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, nil, txInfo, true),
		TxInfo:       txInfo.(*txtypes.DepositNftTxInfo),
	}, nil
}

func (e *DepositNftExecutor) Prepare() error {
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

	_, err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return err
	}

	// Mark the tree states that would be affected in this executor.
	e.MarkNftDirty(txInfo.NftIndex)
	if e.IsCreateAccount {
		err := e.CreateEmptyAccount(txInfo.AccountIndex, l1Address, []int64{types.EmptyAccountAssetId})
		if err != nil {
			return err
		}
	}
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{types.EmptyAccountAssetId}) // Prepare asset 0 for generate an empty tx detail.
	return e.BaseExecutor.Prepare()
}

func (e *DepositNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	bc := e.bc
	txInfo := e.TxInfo

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
	txInfo := e.TxInfo
	bc := e.bc
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

	nft := &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.AccountIndex,
		NftContentHash:      common.Bytes2Hex(txInfo.NftContentHash),
		RoyaltyRate:         txInfo.RoyaltyRate,
		CollectionId:        txInfo.CollectionId,
		NftContentType:      txInfo.NftContentType,
	}
	cacheNft, err := e.bc.StateDB().GetNft(txInfo.NftIndex)
	if err == nil {
		nft.ID = cacheNft.ID
	}
	stateCache := e.bc.StateDB()
	stateCache.SetPendingNft(txInfo.NftIndex, nft)
	stateCache.SetPendingAccount(depositAccount.AccountIndex, depositAccount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *DepositNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeDepositNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.WriteByte(uint8(txInfo.NftContentType))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *DepositNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.NftIndex = e.TxInfo.NftIndex
	e.tx.IsCreateAccount = e.IsCreateAccount
	if e.tx.ToAccountIndex != e.iTxInfo.GetToAccountIndex() || e.IsCreateAccount {
		e.tx.IsPartialUpdate = true
	}
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *DepositNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
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
		L1Address:       depositAccount.L1Address,
		Balance:         baseBalance.String(),
		BalanceDelta:    deltaBalance.String(),
		AccountOrder:    accountOrder,
		Order:           order,
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
			AccountOrder:    accountOrder,
			Nonce:           depositAccount.Nonce,
			CollectionNonce: depositAccount.CollectionNonce,
			PublicKey:       depositAccount.PublicKey,
		})
	}
	// nft info
	order++
	baseNft := types.EmptyNftInfo(txInfo.NftIndex)
	newNft := types.ConstructNftInfo(
		txInfo.NftIndex,
		txInfo.CreatorAccountIndex,
		txInfo.AccountIndex,
		common.Bytes2Hex(txInfo.NftContentHash),
		txInfo.RoyaltyRate,
		txInfo.CollectionId,
		txInfo.NftContentType,
	)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.AccountIndex,
		L1Address:       depositAccount.L1Address,
		Balance:         baseNft.String(),
		BalanceDelta:    newNft.String(),
		AccountOrder:    types.NilAccountOrder,
		Order:           order,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
		PublicKey:       depositAccount.PublicKey,
	})

	return txDetails, nil
}

func (e *DepositNftExecutor) Finalize() error {
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
