package block

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/utils"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByCommitmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetBlockByCommitmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByCommitmentLogic {
	return &GetBlockByCommitmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetBlockByCommitmentLogic) GetBlockByCommitment(req *types.ReqGetBlockByCommitment) (*types.RespGetBlockByCommitment, error) {
	// todo: add your logic here and delete this line
	// query basic block info
	block, err := l.block.GetBlockWithTxsByCommitment(req.BlockCommitment)
	if err != nil {
		logx.Errorf("[GetBlockWithTxsByCommitment] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetBlockByCommitment{
		Block: types.Block{
			BlockCommitment:                 block.BlockCommitment,
			BlockHeight:                     block.BlockHeight,
			StateRoot:                       block.StateRoot,
			PriorityOperations:              block.PriorityOperations,
			PendingOnChainOperationsHash:    block.PendingOnChainOperationsHash,
			PendingOnChainOperationsPubData: block.PendingOnChainOperationsPubData,
			CommittedTxHash:                 block.CommittedTxHash,
			CommittedAt:                     block.BlockHeight,
			VerifiedTxHash:                  block.VerifiedTxHash,
			VerifiedAt:                      block.BlockHeight,
			BlockStatus:                     block.BlockHeight,
		},
	}
	for _, t := range block.Txs {
		tx := utils.GormTx2Tx(t)
		resp.Block.Txs = append(resp.Block.Txs, tx)
	}
	return resp, nil
}
