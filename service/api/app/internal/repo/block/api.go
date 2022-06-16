package block

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
)

type Block interface {
	GetCommitedBlocksCount(ctx context.Context) (count int64, err error)
	GetExecutedBlocksCount(ctx context.Context) (count int64, err error)
	GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error)
}

func New(svcCtx *svc.ServiceContext) Block {
	return &block{
		table: `block`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
