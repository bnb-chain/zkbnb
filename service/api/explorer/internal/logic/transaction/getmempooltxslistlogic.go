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

type GetMempoolTxsListLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Model
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetMempoolTxsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListLogic {
	return &GetMempoolTxsListLogic{
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

func (l *GetMempoolTxsListLogic) GetMempoolTxsList(req *types.ReqGetMempoolTxsList) (*types.RespGetMempoolTxsList, error) {
	mempoolTxs, err := l.mempool.GetMempoolTxs(int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetMempoolTxs] err:%v", err)
		return nil, err
	}
	total, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		logx.Errorf("[GetMempoolTxsTotalCount] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetMempoolTxsList{
		Total: uint32(total),
	}
	for _, tx := range mempoolTxs {
		resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
