package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempooltxdetail"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetmempoolTxsByAccountNameLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	account         account.Model
	memPoolTxDetail mempooltxdetail.Model
	mempool         mempool.Mempool
}

func NewGetmempoolTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmempoolTxsByAccountNameLogic {
	return &GetmempoolTxsByAccountNameLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		account:         account.New(svcCtx),
		memPoolTxDetail: mempooltxdetail.New(svcCtx),
		mempool:         mempool.New(svcCtx),
	}
}

func (l *GetmempoolTxsByAccountNameLogic) GetmempoolTxsByAccountName(req *types.ReqGetmempoolTxsByAccountName) (*types.RespGetmempoolTxsByAccountName, error) {
	resp := &types.RespGetmempoolTxsByAccountName{
		Txs: make([]*types.Tx, 0),
	}
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	mempoolTxDetails, err := l.memPoolTxDetail.GetMemPoolTxDetailByAccountIndex(l.ctx, int64(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetMemPoolTxDetailByAccountIndex] AccountIndex:%v err:%v", account.AccountIndex, err)
		return nil, err
	}
	for _, d := range mempoolTxDetails {
		tx, err := l.mempool.GetMempoolTxByTxId(l.ctx, d.TxId)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxID] TxId:%v, err:%v", d.TxId, err)
			continue
		}
		resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
