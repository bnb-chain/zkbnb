package transaction

import (
	"context"
	"strconv"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	blockdao "github.com/bnb-chain/zkbas/common/model/block"
)

const (
	queryByBlockHeight     = "block_height"
	queryByBlockCommitment = "block_commitment"
)

type GetBlockTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockTxsLogic {
	return &GetBlockTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockTxsLogic) GetBlockTxs(req *types.ReqGetBlockTxs) (resp *types.Txs, err error) {
	var block *blockdao.Block
	switch req.By {
	case queryByBlockHeight:
		blockHeight, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		if blockHeight <= 0 {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for block height")
		}
		block, err = l.svcCtx.MemCache.GetBlockByHeightWithFallback(blockHeight, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByHeight(blockHeight)
		})
	case queryByBlockCommitment:
		block, err = l.svcCtx.MemCache.GetBlockByCommitmentWithFallback(req.Value, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByCommitment(req.Value)
		})
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be height|commitment")
	}

	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.Txs{
		Total: uint32(len(block.Txs)),
		Txs:   make([]*types.Tx, 0),
	}
	for _, t := range block.Txs {
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
