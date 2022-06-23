package block

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

type Block interface {
	GetCommitedBlocksCount() (count int64, err error)
	GetVerifiedBlocksCount() (count int64, err error)
	GetBlockWithTxsByCommitment(BlockCommitment string) (block *table.Block, err error)
	GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlockWithTxsByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlocksList(limit int64, offset int64) (blocks []*table.Block, err error)
	GetBlocksTotalCount() (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Block {
	return &block{
		table: `block`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
