package core

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
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
	var start time.Time
	p.bc.setCurrentBlockTimeStamp()
	defer func() {
		if err := recover(); err != nil {
			if types.IsL2Tx(tx.TxType) {
				expectNonce, err := p.bc.Statedb.GetCommittedNonce(tx.AccountIndex)
				if err != nil {
					p.bc.Statedb.ClearPendingNonceFromRedisCache(tx.AccountIndex)
				} else {
					p.bc.Statedb.SetPendingNonceToRedisCache(tx.AccountIndex, expectNonce-1)
				}
			}
			logx.Severe(err)
			panic(err)
		}
	}()
	defer p.bc.resetCurrentBlockTimeStamp()

	executor, err := executor.NewTxExecutor(p.bc, tx)
	if err != nil {
		return fmt.Errorf("new tx executor failed")
	}
	start = time.Now()
	err = executor.Prepare()
	metrics.ExecuteTxPrepareMetrics.Set(float64(time.Since(start).Milliseconds()))

	if err != nil {
		return err
	}
	start = time.Now()
	err = executor.VerifyInputs(true, true)
	metrics.ExecuteTxVerifyInputsMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		return err
	}
	start = time.Now()
	txDetails, err := executor.GenerateTxDetails()

	metrics.ExecuteGenerateTxDetailsMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		return err
	}
	for _, txDetail := range txDetails {
		txDetail.PoolTxId = tx.ID
		txDetail.BlockHeight = p.bc.currentBlock.BlockHeight
	}
	tx.TxDetails = txDetails
	start = time.Now()
	err = executor.ApplyTransaction()
	metrics.ExecuteTxApplyTransactionMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	start = time.Now()
	err = executor.GeneratePubData()
	metrics.ExecuteTxGeneratePubDataMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	start = time.Now()
	tx, err = executor.GetExecutedTx(false)
	metrics.ExecuteTxGetExecutedTxMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	err = executor.Finalize()
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	tx.CreatedAt = time.Now()
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
	err = executor.VerifyInputs(false, false)
	if err != nil {
		return mappingVerifyInputsErrors(err)
	}
	_, err = executor.GetExecutedTx(true)
	if err != nil {
		return mappingExecutedErrors(err)
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

func mappingExecutedErrors(err error) error {
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
