package block

import (
	"context"
	"fmt"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

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

func (l *GetBlockByCommitmentLogic) GetBlockByCommitment(req *types.ReqGetBlockByCommitment) (resp *types.RespGetBlockByCommitment, err error) {
	// query basic block info
	blockInfo, err := l.block.GetBlockWithTxsByCommitment(req.BlockCommitment)
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlockWithTxsByCommitment]<=>%s", err.Error())
		l.Error(err)
		return
	}

	resp.Block = types.Block{
		BlockHeight:    int32(blockInfo.BlockHeight),
		BlockStatus:    int32(blockInfo.BlockStatus),
		NewAccountRoot: blockInfo.StateRoot,
		CommittedAt:    blockInfo.CommittedAt,
		VerifiedAt:     blockInfo.VerifiedAt,
		// ExecutedAt: block.,
		BlockCommitment: blockInfo.BlockCommitment,
		TxCount:         int64(len(blockInfo.Txs)),
	}

	for _, tx := range blockInfo.Txs {
		resp.Block.Txs = append(resp.Block.Txs, tx.TxHash)
	}

	for _, tx := range blockInfo.Txs {
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
