package block

import (
	"context"
	"fmt"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlocksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetBlocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlocksLogic {
	return &GetBlocksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetBlocksLogic) GetBlocks(req *types.ReqGetBlocks) (resp *types.RespGetBlocks, err error) {
	blocks, err := l.block.GetBlocksList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlocks]<=>%s", err.Error())
		l.Error(err)
		return
	}
	total, err := l.block.GetBlocksTotalCount()
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlocksTotalCount]<=>%s", err.Error())
		l.Error(err)
		return
	}
	resp.Total = uint32(total)

	for _, b := range blocks {
		block := types.Block{
			BlockHeight:    int32(b.BlockHeight),
			BlockStatus:    int32(b.BlockStatus),
			NewAccountRoot: b.StateRoot,
			CommittedAt:    b.CommittedAt,
			VerifiedAt:     b.VerifiedAt,
			// ExecutedAt: block.,
			BlockCommitment: b.BlockCommitment,
			TxCount:         int64(len(b.Txs)),
		}

		for _, tx := range b.Txs {
			block.Txs = append(block.Txs, tx.TxHash)
		}

		for _, tx := range b.Txs {
			block.CommittedTxHash = append(block.CommittedTxHash, &types.TxHash{
				TxHash:    tx.TxHash,
				CreatedAt: tx.CreatedAt.Unix(),
			})

			block.VerifiedTxHash = append(block.VerifiedTxHash, &types.TxHash{
				TxHash:    tx.TxHash,
				CreatedAt: tx.CreatedAt.Unix(),
			})

			block.ExecutedTxHash = append(block.ExecutedTxHash, &types.TxHash{
				TxHash:    tx.TxHash,
				CreatedAt: tx.CreatedAt.Unix(),
			})
		}

		resp.Blocks = append(resp.Blocks, &block)
	}
	return
}
