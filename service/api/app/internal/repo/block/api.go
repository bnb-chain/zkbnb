package block

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Block interface {
	GetCommitedBlocksCount(ctx context.Context) (count int64, err error)
	GetVerifiedBlocksCount(ctx context.Context) (count int64, err error)
	GetBlockWithTxsByCommitment(ctx context.Context, BlockCommitment string) (block *table.Block, err error)
	GetBlockByBlockHeight(ctx context.Context, blockHeight int64) (block *table.Block, err error)
	GetBlockWithTxsByBlockHeight(ctx context.Context, blockHeight int64) (block *table.Block, err error)
	GetBlocksList(ctx context.Context, limit int64, offset int64) (blocks []*table.Block, err error)
	GetBlocksTotalCount(ctx context.Context) (count int64, err error)
	GetCurrentBlockHeight(ctx context.Context) (height int64, err error)
}

func New(svcCtx *svc.ServiceContext) Block {
	return &block{
		table: `block`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
