package core

import (
	"fmt"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type Processor interface {
	Process(tx *tx.Tx) error
}

type CommitProcessor struct {
	bc *BlockChain
	// TODO make it as an option
	trace bool
}

func NewCommitProcessor(bc *BlockChain) Processor {
	return &CommitProcessor{
		bc:    bc,
		trace: true,
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
	if p.trace {
		witness, err := executor.GenerateWitness()
		if err != nil {
			panic(err)
		}
		p.bc.Statedb.Witnesses = append(p.bc.Statedb.Witnesses, witness)
	}
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

	if p.trace {
		// Intermediate state root.
		err := p.bc.Statedb.IntermediateRoot(false)
		if err != nil {
			panic(err)
		}
		p.bc.Statedb.Witnesses[len(p.bc.Statedb.Witnesses)-1].StateRootAfter = hexutils.HexToBytes(p.bc.Statedb.StateRoot)
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
		return err
	}
	err = executor.VerifyInputs()
	if err != nil {
		return types.AppErrInvalidTxField.RefineError(err.Error())
	}

	return nil
}
