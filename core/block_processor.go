package core

import (
	"fmt"

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
	err = executor.VerifyInputs()
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
	err = executor.VerifyInputs()
	if err != nil {
		return types.AppErrInvalidTxField.RefineError(err.Error())
	}

	return nil
}
