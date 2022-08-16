package transaction

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	blockHeight := int64(0)
	switch req.By {
	case "height":
		blockHeight, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for block_height")
		}
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be block_height")
	}

	block, err := l.svcCtx.MemCache.GetBlockByHeightWithFallback(blockHeight, func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetBlockByBlockHeight(blockHeight)
	})
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
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
