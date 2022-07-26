package block

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetBlockByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByBlockHeightLogic {
	return &GetBlockByBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetBlockByBlockHeightLogic) GetBlockByBlockHeight(req *types.ReqGetBlockByBlockHeight) (*types.RespGetBlockByBlockHeight, error) {
	block, err := l.block.GetBlockWithTxsByBlockHeight(l.ctx, int64(req.BlockHeight))
	if err != nil {
		logx.Errorf("[GetBlockWithTxsByBlockHeight] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetBlockByBlockHeight{
		Block: types.Block{
			BlockCommitment:                 block.BlockCommitment,
			BlockHeight:                     block.BlockHeight,
			StateRoot:                       block.StateRoot,
			PriorityOperations:              block.PriorityOperations,
			PendingOnChainOperationsHash:    block.PendingOnChainOperationsHash,
			PendingOnChainOperationsPubData: block.PendingOnChainOperationsPubData,
			CommittedTxHash:                 block.CommittedTxHash,
			CommittedAt:                     block.CommittedAt,
			VerifiedTxHash:                  block.VerifiedTxHash,
			VerifiedAt:                      block.VerifiedAt,
			BlockStatus:                     block.BlockStatus,
		},
	}
	for _, t := range block.Txs {
		tx := utils.GormTx2Tx(t)
		resp.Block.Txs = append(resp.Block.Txs, tx)
	}
	return resp, nil
}
