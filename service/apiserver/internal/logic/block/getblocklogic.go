package block

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	blockdao "github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
		var height int64
		height, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil || height < 0 {
			return nil, types2.AppErrInvalidBlockHeight
		}
		block, err = l.svcCtx.MemCache.GetBlockByHeightWithFallback(height, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByHeight(height)
		})
	case queryByCommitment:
		block, err = l.svcCtx.MemCache.GetBlockByCommitmentWithFallback(req.Value, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByCommitment(req.Value)
		})
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be height|commitment")
	}
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrBlockNotFound
		}
		return nil, types2.AppErrInternal
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
		Size:                            block.BlockSize,
	}
	for _, dbTx := range block.Txs {
		tx := utils.ConvertTx(dbTx)
		tx.L1Address, _ = l.svcCtx.MemCache.GetL1AddressByIndex(tx.AccountIndex)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
