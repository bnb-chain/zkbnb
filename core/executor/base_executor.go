package executor

import (
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

const (
	OfferPerAsset = 128
	TenThousand   = 10000
)

type BaseExecutor struct {
	bc      IBlockchain
	tx      *tx.Tx
	iTxInfo legendTxTypes.TxInfo
}

func (e *BaseExecutor) Prepare() error {
	return nil
}

func (e *BaseExecutor) VerifyInputs() error {
	txInfo := e.iTxInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}
	err = e.bc.VerifyExpiredAt(txInfo.GetExpiredAt())
	if err != nil {
		return err
	}

	from := txInfo.GetFromAccountIndex()
	if from != types.NilTxAccountIndex {
		err = e.bc.VerifyNonce(from, txInfo.GetNonce())
		if err != nil {
			return err
		}

		err = txInfo.VerifySignature(e.bc.StateDB().AccountMap[from].PublicKey)
		if err != nil {
			return err
		}
	}

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
	e.tx.BlockHeight = e.bc.CurrentBlock().BlockHeight
	e.tx.StateRoot = e.bc.StateDB().GetStateRoot()
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *BaseExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}
