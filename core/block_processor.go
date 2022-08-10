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

func (p *CommitProcessor) Process(tx *tx.Tx, stateCache *StateCache) (*tx.Tx, *StateCache, error) {
	executor, err := NewTxExecutor(p.bc, tx)
	if err != nil {
		return tx, stateCache, fmt.Errorf("new tx executor failed: %v", err)
	}

	err = executor.Prepare()
	if err != nil {
		return tx, stateCache, err
	}
	err = executor.VerifyInputs()
	if err != nil {
		return tx, stateCache, err
	}
	stateCache, err = executor.ApplyTransaction(stateCache)
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

	return tx, stateCache, nil
}
