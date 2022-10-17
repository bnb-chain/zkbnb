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
		return mappingPrepareErrors(err)
	}
	err = executor.VerifyInputs(false)
	if err != nil {
		return mappingVerifyInputsErrors(err)
	}

	return nil
}

func mappingPrepareErrors(err error) error {
	switch e := errors.Cause(err).(type) {
	case types.Error:
		return e
	default:
		return types.AppErrInternal
	}
}

func mappingVerifyInputsErrors(err error) error {
	e := errors.Cause(err)
	switch e {
	case txtypes.ErrAccountIndexTooLow, txtypes.ErrAccountIndexTooHigh,
		txtypes.ErrCreatorAccountIndexTooLow, txtypes.ErrCreatorAccountIndexTooHigh,
		txtypes.ErrFromAccountIndexTooLow, txtypes.ErrFromAccountIndexTooHigh,
		txtypes.ErrToAccountIndexTooLow, txtypes.ErrToAccountIndexTooHigh:
		return types.AppErrInvalidAccountIndex
	case txtypes.ErrGasAccountIndexTooLow, txtypes.ErrGasAccountIndexTooHigh:
		return types.AppErrInvalidGasFeeAccount
	case txtypes.ErrGasFeeAssetIdTooLow, txtypes.ErrGasFeeAssetIdTooHigh:
		return types.AppErrInvalidGasFeeAsset
	case txtypes.ErrGasFeeAssetAmountTooLow, txtypes.ErrGasFeeAssetAmountTooHigh:
		return types.AppErrInvalidGasFeeAmount
	case txtypes.ErrNonceTooLow:
		return types.AppErrInvalidNonce
	case txtypes.ErrOfferTypeInvalid:
		return types.AppErrInvalidOfferType
	case txtypes.ErrOfferIdTooLow:
		return types.AppErrInvalidOfferId
	case txtypes.ErrNftIndexTooLow:
		return types.AppErrInvalidNftIndex
	case txtypes.ErrAssetIdTooLow, txtypes.ErrAssetIdTooHigh:
		return types.AppErrInvalidAssetId
	case txtypes.ErrAssetAmountTooLow, txtypes.ErrAssetAmountTooHigh:
		return types.AppErrInvalidAssetAmount
	case txtypes.ErrListedAtTooLow:
		return types.AppErrInvalidListTime
	case txtypes.ErrTreasuryRateTooLow, txtypes.ErrTreasuryRateTooHigh,
		txtypes.ErrCreatorTreasuryRateTooLow, txtypes.ErrCreatorTreasuryRateTooHigh:
		return types.AppErrInvalidTreasuryRate
	case txtypes.ErrCollectionNameTooShort, txtypes.ErrCollectionNameTooLong:
		return types.AppErrInvalidCollectionName
	case txtypes.ErrIntroductionTooLong:
		return types.AppErrInvalidIntroduction
	case txtypes.ErrNftContentHashInvalid:
		return types.AppErrInvalidNftContenthash
	case txtypes.ErrNftCollectionIdTooLow, txtypes.ErrNftCollectionIdTooHigh:
		return types.AppErrInvalidCollectionId
	case txtypes.ErrCallDataHashInvalid:
		return types.AppErrInvalidCallDataHash
	case txtypes.ErrToAccountNameHashInvalid:
		return types.AppErrInvalidToAccountNameHash
	case txtypes.ErrToAddressInvalid:
		return types.AppErrInvalidToAddress
	case txtypes.ErrBuyOfferInvalid:
		return types.AppErrInvalidBuyOffer
	case txtypes.ErrSellOfferInvalid:
		return types.AppErrInvalidSellOffer

	default:
		return types.AppErrInvalidTxField.RefineError(err.Error())
	}
}
