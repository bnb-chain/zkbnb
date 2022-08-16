package block

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetBlocksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlocksLogic {
	return &GetBlocksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlocksLogic) GetBlocks(req *types.ReqGetAll) (*types.Blocks, error) {
	total, err := l.svcCtx.MemCache.GetBlockTotalCountWithFallback(func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetCurrentBlockHeight()
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp := &types.Blocks{
		Blocks: make([]*types.Block, 0),
		Total:  uint32(total),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	blocks, err := l.svcCtx.BlockModel.GetBlocksList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	for _, b := range blocks {
		block := &types.Block{
			BlockCommitment:                 b.BlockCommitment,
			BlockHeight:                     b.BlockHeight,
			StateRoot:                       b.StateRoot,
			PriorityOperations:              b.PriorityOperations,
			PendingOnChainOperationsHash:    b.PendingOnChainOperationsHash,
			PendingOnChainOperationsPubData: b.PendingOnChainOperationsPubData,
			CommittedTxHash:                 b.CommittedTxHash,
			CommittedAt:                     b.CommittedAt,
			VerifiedTxHash:                  b.VerifiedTxHash,
			VerifiedAt:                      b.VerifiedAt,
			BlockStatus:                     b.BlockStatus,
		}
		for _, t := range b.Txs {
			tx := utils.DbTx2Tx(t)
			tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
			block.Txs = append(block.Txs, tx)
		}
		resp.Blocks = append(resp.Blocks, block)
	}
	return resp, nil
}
