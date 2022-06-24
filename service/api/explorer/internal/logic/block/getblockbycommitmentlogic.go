package block

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/utils"


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
	// query basic block info
	block, err := l.block.GetBlockWithTxsByCommitment(req.BlockCommitment)
	if err != nil {
		logx.Errorf("[GetBlockWithTxsByCommitment] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetBlockByCommitment{}
	resp.Block = types.Block{
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
	}
	for _, t := range block.Txs {
		tx := utils.GormTx2Tx(t)
		resp.Block.Txs = append(resp.Block.Txs, tx)
	}
	return resp, nil
}
