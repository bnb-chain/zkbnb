package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type ChangePubKeyExecutor struct {
	BaseExecutor

	TxInfo *txtypes.ChangePubKeyInfo
}

func NewChangePubKeyExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseChangePubKeyTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse change pub key tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &ChangePubKeyExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		TxInfo:       txInfo,
	}, nil
}

func (e *ChangePubKeyExecutor) Prepare() error {
	txInfo := e.TxInfo

	// Mark the tree states that would be affected in this executor.
	e.MarkAccountAssetsDirty(txInfo.AccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})

	err := e.BaseExecutor.Prepare()
	if err != nil {
		return err
	}

	return nil
}

func (e *ChangePubKeyExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	bc := e.bc
	txInfo := e.TxInfo

	fromAccount, err := bc.StateDB().GetAccountByL1Address(txInfo.L1Address)
	if err != nil {
		return types.AppErrAccountNotFound
	}

	// Verify l1 signature.
	if txInfo.GetL1AddressBySignatureInfo() != common.HexToAddress(fromAccount.L1Address) {
		return types.DbErrFailToL1Signature
	}

	if txInfo.AccountIndex != fromAccount.AccountIndex {
		return types.AppErrInvalidAccountIndex
	}

	err = e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	fromFormatAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	if fromFormatAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}
	return nil
}

func (e *ChangePubKeyExecutor) ApplyTransaction() error {
	txInfo := e.TxInfo
	var err error

	fromAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	if err != nil {
		return err
	}
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(txInfo.PubKeyX)
	pk.A.Y.SetBytes(txInfo.PubKeyY)
	fromAccount.PublicKey = common.Bytes2Hex(pk.Bytes())
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++
	fromAccount.Status = account.AccountStatusConfirmed

	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.AccountIndex, fromAccount)
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return e.BaseExecutor.ApplyTransaction()
}

func (e *ChangePubKeyExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeChangePubKey))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.PubKeyX))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.PubKeyY))
	buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.Nonce)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *ChangePubKeyExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *ChangePubKeyExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.AccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 3)

	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(txInfo.PubKeyX)
	pk.A.Y.SetBytes(txInfo.PubKeyY)
	publicKey := common.Bytes2Hex(pk.Bytes())
	// from account
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         types.NilAssetId,
		AssetType:       types.ChangePubKeyType,
		AccountIndex:    txInfo.AccountIndex,
		L1Address:       fromAccount.L1Address,
		Balance:         fromAccount.PublicKey,
		BalanceDelta:    publicKey,
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.PublicKey = publicKey

	// from account gas
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.AccountIndex,
		L1Address:    fromAccount.L1Address,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
		PublicKey:       fromAccount.PublicKey,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		L1Address:    gasAccount.L1Address,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
		).String(),
		Order:        order,
		Nonce:        gasAccount.Nonce,
		AccountOrder: accountOrder,
		IsGas:        true,
		PublicKey:    gasAccount.PublicKey,
	})
	return txDetails, nil
}
