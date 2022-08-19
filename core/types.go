package core

import (
	"errors"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

type Processor interface {
	Process(tx *tx.Tx) error
}

type TxExecutor interface {
	Prepare() error
	VerifyInputs() error
	ApplyTransaction() error
	GeneratePubData() error
	UpdateTrees() error
	GetExecutedTx() (*tx.Tx, error)
	GenerateTxDetails() ([]*tx.TxDetail, error)
}

func NewTxExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	switch tx.TxType {
	case commonTx.TxTypeRegisterZns:
		return NewRegisterZnsExecutor(bc, tx)
	case commonTx.TxTypeCreatePair:
		return NewCreatePairExecutor(bc, tx)
	case commonTx.TxTypeUpdatePairRate:
		return NewUpdatePairRateExecutor(bc, tx)
	case commonTx.TxTypeDeposit:
		return NewDepositExecutor(bc, tx)
	case commonTx.TxTypeDepositNft:
		return NewDepositNftExecutor(bc, tx)
	case commonTx.TxTypeTransfer:
		return NewTransferExecutor(bc, tx)
	case commonTx.TxTypeSwap:
		return NewSwapExecutor(bc, tx)
	case commonTx.TxTypeAddLiquidity:
		return NewAddLiquidityExecutor(bc, tx)
	case commonTx.TxTypeRemoveLiquidity:
		return NewRemoveLiquidityExecutor(bc, tx)
	case commonTx.TxTypeWithdraw:
		return NewWithdrawExecutor(bc, tx)
	case commonTx.TxTypeCreateCollection:
		return NewCreateCollectionExecutor(bc, tx)
	case commonTx.TxTypeMintNft:
		return NewMintNftExecutor(bc, tx)
	case commonTx.TxTypeTransferNft:
		return NewTransferNftExecutor(bc, tx)
	case commonTx.TxTypeAtomicMatch:
		return NewAtomicMatchExecutor(bc, tx)
	case commonTx.TxTypeCancelOffer:
		return NewCancelOfferExecutor(bc, tx)
	case commonTx.TxTypeWithdrawNft:
		return NewWithdrawNftExecutor(bc, tx)
	case commonTx.TxTypeFullExit:
		return NewFullExitExecutor(bc, tx)
	case commonTx.TxTypeFullExitNft:
		return NewFullExitNftExecutor(bc, tx)
	}

	return nil, errors.New("unsupported tx type")
}
