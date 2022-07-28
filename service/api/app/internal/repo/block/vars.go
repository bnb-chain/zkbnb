package block

const (
	_ = iota
	StatusPending
	StatusCommitted
	StatusVerified
	StatusExecuted
)

const (
	DetailTableName   = `block_detail`
	BlockStatusColumn = "block_status"
	CommittedAtColumn = "committed_at"
	VerifiedAtColumn  = "verified_at"
	ExecutedAtColumn  = "executed_at"
)
