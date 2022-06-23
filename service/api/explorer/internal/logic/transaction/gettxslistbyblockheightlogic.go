package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByBlockHeightLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Model
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetTxsListByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByBlockHeightLogic {
	return &GetTxsListByBlockHeightLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		tx:        tx.New(svcCtx),
		block:     block.New(svcCtx),
		account:   account.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetTxsListByBlockHeightLogic) GetTxsListByBlockHeight(req *types.ReqGetTxsListByBlockHeight) (*types.RespGetTxsListByBlockHeight, error) {
	block, err := l.block.GetBlockWithTxsByBlockHeight(int64(req.BlockHeight))
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
