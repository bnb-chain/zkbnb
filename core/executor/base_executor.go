package executor

import (
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
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
	if from != types.NilAccountIndex {
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

func (e *BaseExecutor) GetExecutedTx() (*tx.Tx, error) {
	e.tx.BlockHeight = e.bc.CurrentBlock().BlockHeight
	e.tx.TxStatus = tx.StatusSuccess
	e.tx.TxIndex = int64(len(e.bc.StateDB().Txs))
	return e.tx, nil
}

func (e *BaseExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}
