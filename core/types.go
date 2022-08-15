package core

import (
	"fmt"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

type Processor interface {
	Process(tx *tx.Tx) (*tx.Tx, error)
}

type TxExecutor interface {
	Prepare() error
	VerifyInputs() error
	ApplyTransaction() error
	GeneratePubData() error
	UpdateTrees() error
	GetExecutedTx() (*tx.Tx, error)
	GenerateTxDetails() []*tx.TxDetail
}

func NewTxExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	switch tx.TxType {
	case commonTx.TxTypeTransfer:
		return NewTransferExecutor(bc, tx)
	}

	return nil, fmt.Errorf("unknow tx type")
}
