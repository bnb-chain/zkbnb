package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/checker"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/utils"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/txdetail"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.Model
	tx        tx.Model
	globalRpc globalrpc.GlobalRPC
	mempool   mempool.Mempool
	block     block.Block
	txDetail  txdetail.Model
}

func NewGetTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountNameLogic {
	return &GetTxsByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx),
		globalRpc: globalrpc.New(svcCtx, ctx),
		tx:        tx.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		block:     block.New(svcCtx),
		txDetail:  txdetail.New(svcCtx),
	}
}

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (*types.RespGetTxsByAccountName, error) {
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	txIds, err := l.txDetail.GetTxIdsByAccountIndex(l.ctx, int64(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetTxDetailByAccountIndex] err:%v", err)
		return nil, err
	}
	logx.Errorf("[GetTxDetailByAccountIndex] err:%v", txIds)

	resp := &types.RespGetTxsByAccountName{
		Total: uint32(len(txIds)),
		Txs:   make([]*types.Tx, 0),
	}
	if checker.CheckOfferset(req.Offset, resp.Total) {
		return nil, errcode.ErrInvalidParam
	}
	end := req.Offset + req.Limit
	if resp.Total < (req.Offset + req.Limit) {
		end = resp.Total
	}
	for _, id := range txIds[req.Offset:end] {
		tx, err := l.tx.GetTxByTxID(l.ctx, id)
		if err != nil {
			logx.Errorf("[GetTxByTxID] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
	}
	return resp, nil

}
