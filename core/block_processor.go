package core

import (
	"fmt"

	"github.com/bnb-chain/zkbas/common/model/tx"
)

type CommitProcessor struct {
	bc *BlockChain
}

func NewCommitProcessor(bc *BlockChain) Processor {
	return &CommitProcessor{
		bc: bc,
	}
}

func (p *CommitProcessor) Process(tx *tx.Tx) (*tx.Tx, error) {
	executor, err := NewTxExecutor(p.bc, tx)
	if err != nil {
		return tx, fmt.Errorf("new tx executor failed: %v", err)
	}

	err = executor.Prepare()
	if err != nil {
		return tx, err
	}
	err = executor.VerifyInputs()
	if err != nil {
		return tx, err
	}
	txDetails, err := executor.GenerateTxDetails()
	if err != nil {
		return tx, err
	}
	tx.TxDetails = txDetails
	err = executor.ApplyTransaction()
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

	p.bc.stateCache.txs = append(p.bc.stateCache.txs, tx)
	p.bc.stateCache.stateRoot = tx.StateRoot

	return tx, nil
}
