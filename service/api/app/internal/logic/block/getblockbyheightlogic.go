package block

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockByHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByHeightLogic {
	return &GetBlockByHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockByHeightLogic) GetBlockByHeight(req *types.ReqGetBlockByHeight) (resp *types.RespGetBlockByHeight, err error) {
	block, err := l.svcCtx.MemCache.GetBlockByHeightWithFallback(int64(req.BlockHeight), func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetBlockByBlockHeight(int64(req.BlockHeight))
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetBlockByHeight{
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
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Block.Txs = append(resp.Block.Txs, tx)
	}
	return resp, nil
}
