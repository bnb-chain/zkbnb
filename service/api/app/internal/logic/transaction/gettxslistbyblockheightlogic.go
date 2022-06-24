package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/utils"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetTxsListByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByBlockHeightLogic {
	return &GetTxsListByBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetTxsListByBlockHeightLogic) GetTxsListByBlockHeight(req *types.ReqGetTxsListByBlockHeight) (*types.RespGetTxsListByBlockHeight, error) {
	block, err := l.block.GetBlockWithTxsByBlockHeight(l.ctx, int64(req.BlockHeight))
	if err != nil {
		logx.Errorf("[GetBlockByBlockHeight] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetTxsListByBlockHeight{
		Total: uint32(len(block.Txs)),
		Txs:   make([]*types.Tx, 0),
	}
	for _, t := range block.Txs {
		tx := utils.GormTx2Tx(t)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
