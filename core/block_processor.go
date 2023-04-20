package core

import (
	"fmt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"math/big"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type Processor interface {
	PreProcess(tx *tx.Tx, accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool)
	Process(tx *tx.Tx) error
}

type ProcessorForDesert interface {
	PreProcess(txInfo txtypes.TxInfo, accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool)
	Process(txInfo txtypes.TxInfo) error
}

type CommitProcessor struct {
	bc *BlockChain
}

func NewCommitProcessor(bc *BlockChain) Processor {
	return &CommitProcessor{
		bc: bc,
	}
}

func (p *CommitProcessor) PreProcess(tx *tx.Tx, accountMap map[int64]bool, nftMap map[int64]bool, addressMap map[string]bool) {
	executor, err := executor.NewTxExecutor(p.bc, tx)
	if err == nil {
		executor.PreLoadAccountAndNft(accountMap, nftMap, addressMap)
	}
}

func (p *CommitProcessor) Process(tx *tx.Tx) (err error) {
	var start time.Time
	p.bc.setCurrentBlockTimeStamp()
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = fmt.Errorf("failed to recover commit processor, %v", recoverErr)
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
		return fmt.Errorf("failed to apply transaction, %v", err)
	}
	tx.IsNonceChanged = true

	start = time.Now()
	err = executor.GeneratePubData()
	metrics.ExecuteTxGeneratePubDataMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to generate PubData, %v", err)
	}
	start = time.Now()
	tx, err = executor.GetExecutedTx(false)
	metrics.ExecuteTxGetExecutedTxMetrics.Set(float64(time.Since(start).Milliseconds()))
	if err != nil {
		return err
	}
	err = executor.Finalize()
	if err != nil {
		return fmt.Errorf("failed to get executed transaction, %v", err)
	}
	tx.CreatedAt = time.Now()
	p.bc.Statedb.Txs = append(p.bc.Statedb.Txs, tx)
	metrics.TxTypeCommitRunCounter.WithLabelValues(strconv.Itoa(executor.GetTxInfo().GetTxType())).Inc()
	fee, isValid := new(big.Int).SetString(tx.GasFee, 10)
	if isValid {
		metrics.TxTypeCommitRevenueCounter.WithLabelValues(strconv.Itoa(executor.GetTxInfo().GetTxType())).Add(common2.GetFeeFromWei(fee))
		metrics.GasAccountRevenueCounter.Add(common2.GetFeeFromWei(fee))
	} else {
		metrics.TxTypeCommitRevenueCounter.WithLabelValues(strconv.Itoa(executor.GetTxInfo().GetTxType())).Add(0)
		metrics.GasAccountRevenueCounter.Add(0)
	}
	return nil
}

type APIProcessor struct {
	bc *BlockChain
}

type DesertProcessor struct {
	bc *BlockChain
}

func NewAPIProcessor(bc *BlockChain) Processor {
	return &APIProcessor{
		bc: bc,
	}
}

func NewDesertProcessor(bc *BlockChain) ProcessorForDesert {
	return &DesertProcessor{
		bc: bc,
	}
}

func (p *APIProcessor) PreProcess(tx *tx.Tx, accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
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
		logx.Error("fail to VerifyInput:", err)
		return mappingVerifyInputsErrors(err)
	}
	_, err = executor.GetExecutedTx(true)
	if err != nil {
		logx.Error("fail to GetExecutedTx:", err)
		return mappingExecutedErrors(err)
	}
	return nil
}

func (p *DesertProcessor) PreProcess(txInfo txtypes.TxInfo, accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
	executor, err := executor.NewTxExecutorForDesert(p.bc, txInfo)
	if err == nil {
		executor.PreLoadAccountAndNft(accountIndexMap, nftIndexMap, addressMap)
	}
}

func (p *DesertProcessor) Process(txInfo txtypes.TxInfo) error {
	executor, err := executor.NewTxExecutorForDesert(p.bc, txInfo)
	if err != nil {
		return fmt.Errorf("new tx executor failed")
	}

	err = executor.Prepare()
	if err != nil {
		logx.Error("fail to prepare:", err)
		return mappingPrepareErrors(err)
	}

	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}

	err = executor.Finalize()
	if err != nil {
		return err
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
	case txtypes.ErrProtocolRateTooLow, txtypes.ErrProtocolRateTooHigh,
		txtypes.ErrChannelRateTooLow, txtypes.ErrChannelRateTooHigh,
		txtypes.ErrRoyaltyRateTooLow, txtypes.ErrRoyaltyRateTooHigh:
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
		return formatVerifyInputsErrors(err)
	}
}

func formatVerifyInputsErrors(err error) error {
	if _, ok := err.(types.Error); ok {
		return err
	}
	return types.AppErrInvalidTxField.RefineError(err)
}
