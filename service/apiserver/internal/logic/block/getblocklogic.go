package block

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	blockdao "github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

const (
	queryByHeight     = "height"
	queryByCommitment = "commitment"
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
	var block *blockdao.Block
	switch req.By {
	case queryByHeight:
		blockHeight, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, types2.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		if blockHeight < 0 {
			return nil, types2.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		block, err = l.svcCtx.MemCache.GetBlockByHeightWithFallback(blockHeight, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByHeight(blockHeight)
		})
	case queryByCommitment:
		block, err = l.svcCtx.MemCache.GetBlockByCommitmentWithFallback(req.Value, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByCommitment(req.Value)
		})
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be height|commitment")
	}

	resp = &types.Block{
		Commitment:                      block.BlockCommitment,
		Height:                          block.BlockHeight,
		StateRoot:                       block.StateRoot,
		PriorityOperations:              block.PriorityOperations,
		PendingOnChainOperationsHash:    block.PendingOnChainOperationsHash,
		PendingOnChainOperationsPubData: block.PendingOnChainOperationsPubData,
		CommittedTxHash:                 block.CommittedTxHash,
		CommittedAt:                     block.CommittedAt,
		VerifiedTxHash:                  block.VerifiedTxHash,
		VerifiedAt:                      block.VerifiedAt,
		Status:                          block.BlockStatus,
	}
	for _, t := range block.Txs {
		tx := utils.DbtxTx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
