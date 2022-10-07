package core

import (
	"fmt"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type Processor interface {
	Process(tx *tx.Tx) error
}

type CommitProcessor struct {
	bc *BlockChain
}

func NewCommitProcessor(bc *BlockChain) Processor {
	return &CommitProcessor{
		bc: bc,
	}
}

func (p *CommitProcessor) Process(tx *tx.Tx) error {
	p.bc.setCurrentBlockTimeStamp()
	defer p.bc.resetCurrentBlockTimeStamp()

	executor, err := executor.NewTxExecutor(p.bc, tx)
	if err != nil {
		return fmt.Errorf("new tx executor failed")
	}

	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.VerifyInputs(true)
	if err != nil {
		return err
	}
	txDetails, err := executor.GenerateTxDetails()
	if err != nil {
		return err
	}
	tx.TxDetails = txDetails
	err = executor.ApplyTransaction()
	if err != nil {
		panic(err)
	}
	err = executor.GeneratePubData()
	if err != nil {
		panic(err)
	}
	tx, err = executor.GetExecutedTx()
	if err != nil {
		panic(err)
	}

	p.bc.Statedb.Txs = append(p.bc.Statedb.Txs, tx)

	return nil
}

type APIProcessor struct {
	bc *BlockChain
}

func NewAPIProcessor(bc *BlockChain) Processor {
	return &APIProcessor{
		bc: bc,
	}
}

func (p *APIProcessor) Process(tx *tx.Tx) error {
	executor, err := executor.NewTxExecutor(p.bc, tx)
	if err != nil {
		return fmt.Errorf("new tx executor failed")
	}

	err = executor.Prepare()
	if err != nil {
		logx.Error("fail to prepare:", err)
		return types.AppErrInternal
	}
	err = executor.VerifyInputs(false)
	if err != nil {
		return mappingAPIErrors(err)
	}

	return nil
}

func mappingAPIErrors(err error) error {
	e := errors.Cause(err)
	switch e {
	case txtypes.ErrAccountIndexTooLow, txtypes.ErrAccountIndexTooHigh,
		txtypes.ErrCreatorAccountIndexTooLow, txtypes.ErrCreatorAccountIndexTooHigh,
		txtypes.ErrFromAccountIndexTooLow, txtypes.ErrFromAccountIndexTooHigh,
		txtypes.ErrToAccountIndexTooLow, txtypes.ErrToAccountIndexTooHigh:
		return types.AppErrTxInvalidAccountIndex
	case txtypes.ErrGasAccountIndexTooLow, txtypes.ErrGasAccountIndexTooHigh:
		return types.AppErrTxInvalidGasFeeAccount
	case txtypes.ErrGasFeeAssetIdTooLow, txtypes.ErrGasFeeAssetIdTooHigh:
		return types.AppErrTxInvalidGasFeeAsset
	case txtypes.ErrGasFeeAssetAmountTooLow, txtypes.ErrGasFeeAssetAmountTooHigh:
		return types.AppErrTxInvalidGasFeeAmount
	case txtypes.ErrNonceTooLow:
		return types.AppErrTxInvalidNonce
	case txtypes.ErrOfferTypeInvalid:
		return types.AppErrTxInvalidOfferType
	case txtypes.ErrOfferIdTooLow:
		return types.AppErrTxInvalidOfferId
	case txtypes.ErrNftIndexTooLow:
		return types.AppErrTxInvalidNftIndex
	case txtypes.ErrAssetIdTooLow, txtypes.ErrAssetIdTooHigh,
		txtypes.ErrAssetAIdTooLow, txtypes.ErrAssetAIdTooHigh:
		return types.AppErrTxInvalidAssetId
	case txtypes.ErrAssetAmountTooLow, txtypes.ErrAssetAmountTooHigh,
		txtypes.ErrAssetAAmountTooLow, txtypes.ErrAssetAAmountTooHigh,
		txtypes.ErrAssetBAmountTooLow, txtypes.ErrAssetBAmountTooHigh:
		return types.AppErrTxInvalidAssetAmount
	case txtypes.ErrListedAtTooLow:
		return types.AppErrTxInvalidListTime
	case txtypes.ErrTreasuryRateTooLow, txtypes.ErrTreasuryRateTooHigh,
		txtypes.ErrCreatorTreasuryRateTooLow, txtypes.ErrCreatorTreasuryRateTooHigh:
		return types.AppErrTxInvalidTreasuryRate
	case txtypes.ErrCollectionNameTooShort, txtypes.ErrCollectionNameTooLong:
		return types.AppErrTxInvalidCollectionName
	case txtypes.ErrIntroductionTooLong:
		return types.AppErrTxInvalidIntroduction
	case txtypes.ErrNftContentHashInvalid:
		return types.AppErrTxInvalidNftContenthash
	case txtypes.ErrNftCollectionIdTooLow, txtypes.ErrNftCollectionIdTooHigh:
		return types.AppErrTxInvalidCollectionId
	case txtypes.ErrPairIndexTooLow, txtypes.ErrPairIndexTooHigh:
		return types.AppErrTxInvalidPairIndex
	case txtypes.ErrLpAmountTooLow, txtypes.ErrLpAmountTooLow:
		return types.AppErrTxInvalidLpAmount
	case txtypes.ErrCallDataHashInvalid:
		return types.AppErrTxInvalidCallDataHash
	case txtypes.ErrToAccountNameHashInvalid:
		return types.AppErrTxInvalidToAccountNameHash
	case txtypes.ErrAssetAMinAmountTooLow, txtypes.ErrAssetAMinAmountTooHigh,
		txtypes.ErrAssetBMinAmountTooLow, txtypes.ErrAssetBMinAmountTooHigh:
		return types.AppErrTxInvalidAssetMinAmount
	case txtypes.ErrToAddressInvalid:
		return types.AppErrTxInvalidToAddress
	case txtypes.ErrBuyOfferInvalid:
		return types.AppErrTxInvalidBuyOffer
	case txtypes.ErrSellOfferInvalid:
		return types.AppErrTxInvalidSellOffer

	default:
		return types.AppErrInvalidTxField.RefineError(err.Error())
	}
}
