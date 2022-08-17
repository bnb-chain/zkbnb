package block

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetBlockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockLogic {
	return &GetBlockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockLogic) GetBlock(req *types.ReqGetBlock) (resp *types.Block, err error) {
	var block *block.Block
	switch req.By {
	case "height":
		blockHeight, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		if blockHeight <= 0 {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		block, err = l.svcCtx.MemCache.GetBlockByHeightWithFallback(blockHeight, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByBlockHeight(blockHeight)
		})
	case "commitment":
		block, err = l.svcCtx.MemCache.GetBlockByCommitmentWithFallback(req.Value, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByCommitment(req.Value)
		})
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be height|commitment")
	}

	resp = &types.Block{
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
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
