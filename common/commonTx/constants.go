package commonTx

const (
	// TODO
	TxTypeNoOp = iota
	TxTypeDeposit
	TxTypeTransfer
	TxTypeSwap
	TxTypeAddLiquidity
	TxTypeRemoveLiquidity
	TxTypeWithdraw
)

const (
	TxPending = iota
	TxSuccess
	TxFail
)

const (
	L2TxChainId = -1
)
