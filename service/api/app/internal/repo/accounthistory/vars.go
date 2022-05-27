package account

const (
	Fail = iota
	NoRegisteredNoActive
	RegisteredNoActive
	Active
)

const (
	AccountTableName         = `account`
	AccountRegisterTableName = `account_register`
	AccountHistoryTableName  = `account_history`
)

const (
	AccountHistoryPending = iota
	AccountHistoryConfirmed
)
