package core

import (
	"fmt"

	"github.com/bnb-chain/zkbas/core/executor"
	"github.com/bnb-chain/zkbas/dao/tx"
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
	err = executor.UpdateTrees()
	if err != nil {
		panic(err)
	}
	tx, err = executor.GetExecutedTx()
	if err != nil {
		panic(err)
	}

	p.bc.Statedb.Txs = append(p.bc.Statedb.Txs, tx)
	p.bc.Statedb.StateRoot = tx.StateRoot

	return nil
}
