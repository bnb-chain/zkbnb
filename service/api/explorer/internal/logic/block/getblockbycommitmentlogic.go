package block

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByCommitmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockByCommitmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByCommitmentLogic {
	return &GetBlockByCommitmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockByCommitmentLogic) GetBlockByCommitment(req *types.ReqGetBlockByCommitment) (resp *types.RespGetBlockByCommitment, err error) {
	// query basic block info
	block, err := l.svcCtx.Block.GetBlockWithTxsByCommitment(req.BlockCommitment)
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlockWithTxsByCommitment]<=>%s", err.Error())
		l.Error(err)
		return
	}

	resp.Block = types.Block{
		BlockHeight:    int32(block.BlockHeight),
		BlockStatus:    int32(block.BlockStatus),
		NewAccountRoot: block.StateRoot,
		CommittedAt:    block.CommittedAt,
		VerifiedAt:     block.VerifiedAt,
		// ExecutedAt: block.,
		BlockCommitment: block.BlockCommitment,
		TxCount:         int64(len(block.Txs)),
	}

	for _, tx := range block.Txs {
		resp.Block.Txs = append(resp.Block.Txs, tx.TxHash)
	}

	for _, tx := range block.Txs {
		resp.Block.CommittedTxHash = append(resp.Block.CommittedTxHash, &types.TxHash{
			TxHash:    tx.TxHash,
			CreatedAt: tx.CreatedAt.Unix(),
		})

		resp.Block.VerifiedTxHash = append(resp.Block.VerifiedTxHash, &types.TxHash{
			TxHash:    tx.TxHash,
			CreatedAt: tx.CreatedAt.Unix(),
		})

		resp.Block.ExecutedTxHash = append(resp.Block.ExecutedTxHash, &types.TxHash{
			TxHash:    tx.TxHash,
			CreatedAt: tx.CreatedAt.Unix(),
		})
	}
	return
}
