package block

import (
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"gorm.io/gorm"
)

type BlockDetailInfo struct {
	gorm.Model
	ChainId               int64 `gorm:"index"`
	VerifiedTxHash        string
	VerifiedAt            int64
	CommittedTxHash       string
	CommittedAt           int64
	ExecutedTxHash        string
	ExecutedAt            int64
	BlockPk               int64      `gorm:"index"`
	Txs                   []*tx.TxDB `gorm:"foreignkey:BlockDetailPk"`
	OnChainPublicData     []byte
	OnChainOpsMerkleProof string
	OnChainOpsCount       int
}

type BlockInfo struct {
	gorm.Model
	BlockCommitment string
	BlockHeight     int64
	BlockStatus     int64
	OnChainOpsRoot  string
	AccountRoot     string
	CommittedAt     int64
	VerifiedAt      int64
	ExecutedAt      int64
	BlockDetails    []*BlockDetailInfo `gorm:"foreignkey:BlockPk"`
	Txs             []*tx.TxDB         `gorm:"foreignkey:BlockId"`
}
