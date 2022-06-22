package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempooldetail"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetmempoolTxsByAccountNameLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	account       account.AccountModel
	mempool       mempool.Mempool
	mempooldetail mempooldetail.Model
}

func NewGetmempoolTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmempoolTxsByAccountNameLogic {
	return &GetmempoolTxsByAccountNameLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		account:       account.New(svcCtx),
		mempooldetail: mempooldetail.New(svcCtx),
		mempool:       mempool.New(svcCtx),
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
	mempoolTxDetails, err := l.mempooldetail.GetMempoolTxDetailByAccountIndex(int64(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetMempoolTxDetailByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, d := range mempoolTxDetails {
		// loop run GetMempoolTxByTxID to add cache with txID
		tx, err := l.mempool.GetMempoolTxByTxID(d.TxId)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxID] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
