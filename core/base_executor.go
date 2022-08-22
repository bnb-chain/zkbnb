package core

import (
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

const (
	OfferPerAsset = 128
	TenThousand   = 10000
)

type BaseExecutor struct {
	bc      *BlockChain
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
	err = e.bc.verifyExpiredAt(txInfo.GetExpiredAt())
	if err != nil {
		return err
	}

	from := txInfo.GetFromAccountIndex()
	if from != commonConstant.NilTxAccountIndex {
		err = e.bc.verifyNonce(from, txInfo.GetNonce())
		if err != nil {
			return err
		}

		err = txInfo.VerifySignature(e.bc.accountMap[from].PublicKey)
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
	e.tx.BlockHeight = e.bc.currentBlock.BlockHeight
	e.tx.StateRoot = e.bc.getStateRoot()
	e.tx.TxStatus = tx.StatusPending
	return e.tx, nil
}

func (e *BaseExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}
