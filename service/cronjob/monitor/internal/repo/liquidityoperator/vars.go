package liquidityoperator

const (
	Fail = iota
	NoRegisteredNoActive
	RegisteredNoActive
	Active
)

const (
	AccountHistoryPending = iota
	AccountHistoryConfirmed
)
