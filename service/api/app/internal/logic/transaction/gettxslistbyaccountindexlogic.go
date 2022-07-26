package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/txdetail"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByAccountIndexLogic struct {
	logx.Logger
	ctx      context.Context
	svcCtx   *svc.ServiceContext
	txDetail txdetail.Model
	tx       tx.Model
}

func NewGetTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByAccountIndexLogic {
	return &GetTxsListByAccountIndexLogic{
		Logger:   logx.WithContext(ctx),
		ctx:      ctx,
		svcCtx:   svcCtx,
		txDetail: txdetail.New(svcCtx),
		tx:       tx.New(svcCtx),
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
		tx, err := l.tx.GetTxByTxID(l.ctx, d.TxId)
		if err != nil {
			logx.Errorf("[GetTxByTxID] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
	}
	return resp, nil
}
