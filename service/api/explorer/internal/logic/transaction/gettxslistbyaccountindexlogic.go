package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/txdetail"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByAccountIndexLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Model
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
	txDetail  txdetail.Model
}

func NewGetTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByAccountIndexLogic {
	return &GetTxsListByAccountIndexLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		tx:        tx.New(svcCtx),
		block:     block.New(svcCtx),
		account:   account.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
		txDetail:  txdetail.New(svcCtx),
	}
}

func (l *GetTxsListByAccountIndexLogic) GetTxsListByAccountIndex(req *types.ReqGetTxsListByAccountIndex) (*types.RespGetTxsListByAccountIndex, error) {
	resp := &types.RespGetTxsListByAccountIndex{
		Txs: make([]*types.Tx, 0),
	}
	txDetails, err := l.txDetail.GetTxDetailByAccountIndex(l.ctx, int64(req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetTxDetailByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, d := range txDetails {
		// loop run GetMempoolTxByTxID to add cache with txID
		tx, err := l.tx.GetTxByTxID(d.TxId)
		if err != nil {
			logx.Errorf("[GetTxByTxID] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
	}
	return resp, nil
}
