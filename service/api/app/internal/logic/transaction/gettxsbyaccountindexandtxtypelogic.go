package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempooltxdetail"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/txdetail"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	tx              tx.Model
	globalRPC       globalrpc.GlobalRPC
	block           block.Block
	mempool         mempool.Mempool
	txDetail        txdetail.Model
	memPoolTxDetail mempooltxdetail.Model
}

func NewGetTxsByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexAndTxTypeLogic {
	return &GetTxsByAccountIndexAndTxTypeLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		tx:              tx.New(svcCtx),
		globalRPC:       globalrpc.New(svcCtx, ctx),
		block:           block.New(svcCtx),
		mempool:         mempool.New(svcCtx),
		txDetail:        txdetail.New(svcCtx),
		memPoolTxDetail: mempooltxdetail.New(svcCtx),
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (*types.RespGetTxsByAccountIndexAndTxType, error) {
	txDetails, err := l.txDetail.GetTxDetailByAccountIndex(l.ctx, int64(req.AccountIndex))
	if err != nil {
		logx.Error("[GetTxDetailByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetTxsByAccountIndexAndTxType{
		Txs: make([]*types.Tx, 0),
	}
	for _, txDetail := range txDetails {
		tx, err := l.tx.GetTxByTxID(l.ctx, txDetail.TxId)
		if err != nil {
			logx.Errorf("[GetTxByTxID] err:%v", err)
			return nil, err
		}
		if tx.TxType == int64(req.TxType) {
			resp.Total = resp.Total + 1
			resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
		}
	}
	memPoolTxDetails, err := l.memPoolTxDetail.GetMemPoolTxDetailByAccountIndex(l.ctx, int64(req.AccountIndex))
	if err != nil {
		logx.Error("[GetMemPoolTxDetailByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, txDetail := range memPoolTxDetails {
		tx, err := l.mempool.GetMempoolTxByTxId(l.ctx, txDetail.TxId)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxId] err:%v", err)
			return nil, err
		}
		if tx.TxType == int64(req.TxType) {
			resp.Total = resp.Total + 1
			resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
		}
	}
	return resp, nil
}
