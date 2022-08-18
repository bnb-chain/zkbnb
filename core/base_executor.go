package core

import "github.com/bnb-chain/zkbas/common/model/tx"

type BaseExecutor struct {
	bc *BlockChain
	tx *tx.Tx
}

func NewBaseExecutor(bc *BlockChain, tx *tx.Tx) TxExecutor {
	return &BaseExecutor{
		bc: bc,
		tx: tx,
	}
}

func (e *BaseExecutor) Prepare() error {
	return nil
}

func (e *BaseExecutor) VerifyInputs() error {
	return nil
}

func (e *BaseExecutor) ApplyTransaction() error {
	return nil
}

func (e *BaseExecutor) GeneratePubData() error {
	return nil
}

func (e *BaseExecutor) UpdateTrees() error {
	return nil
}

func (e *BaseExecutor) GetExecutedTx() (*tx.Tx, error) {
	e.tx.BlockHeight = e.bc.currentBlock.BlockHeight
	e.tx.StateRoot = e.bc.getStateRoot()
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *BaseExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}
