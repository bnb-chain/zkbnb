package core

import (
	"fmt"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

type (
	TransferTxInfo         = legendTxTypes.TransferTxInfo
	SwapTxInfo             = legendTxTypes.SwapTxInfo
	AddLiquidityTxInfo     = legendTxTypes.AddLiquidityTxInfo
	RemoveLiquidityTxInfo  = legendTxTypes.RemoveLiquidityTxInfo
	WithdrawTxInfo         = legendTxTypes.WithdrawTxInfo
	CreateCollectionTxInfo = legendTxTypes.CreateCollectionTxInfo
	MintNftTxInfo          = legendTxTypes.MintNftTxInfo
	TransferNftTxInfo      = legendTxTypes.TransferNftTxInfo
	OfferTxInfo            = legendTxTypes.OfferTxInfo
	AtomicMatchTxInfo      = legendTxTypes.AtomicMatchTxInfo
	CancelOfferTxInfo      = legendTxTypes.CancelOfferTxInfo
	WithdrawNftTxInfo      = legendTxTypes.WithdrawNftTxInfo
)

type Processor interface {
	Process(tx *tx.Tx, stateCache *StateCache) (*tx.Tx, *StateCache, error)
}

type TxExecutor interface {
	Prepare() error
	VerifyInputs() error
	ApplyTransaction(stateCache *StateCache) (*StateCache, error)
	GeneratePubData(stateCache *StateCache) (*StateCache, error)
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
