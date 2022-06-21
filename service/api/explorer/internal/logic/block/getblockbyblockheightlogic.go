package block

import (
	"context"
	"fmt"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	block block.Block
}

func NewGetBlockByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByBlockHeightLogic {
	return &GetBlockByBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetBlockByBlockHeightLogic) GetBlockByBlockHeight(req *types.ReqGetBlockByBlockHeight) (resp *types.RespGetBlockByBlockHeight, err error) {
	// query basic block info
	block, err := l.block.GetBlockWithTxsByBlockHeight(int64(req.BlockHeight))
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlockByBlockHeight]<=>%s", err.Error())
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
