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

type GetMempoolTxsListByPublicKeyLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	account       account.AccountModel
	mempool       mempool.Mempool
	mempooldetail mempooldetail.Model
}

func NewGetMempoolTxsListByPublicKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListByPublicKeyLogic {
	return &GetMempoolTxsListByPublicKeyLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		account:       account.New(svcCtx),
		mempooldetail: mempooldetail.New(svcCtx),
		mempool:       mempool.New(svcCtx),
	}
}

func (l *GetMempoolTxsListByPublicKeyLogic) GetMempoolTxsListByPublicKey(req *types.ReqGetMempoolTxsListByPublicKey) (*types.RespGetMempoolTxsListByPublicKey, error) {
	resp := &types.RespGetMempoolTxsListByPublicKey{
		Txs: make([]*types.Tx, 0),
	}
	account, err := l.account.GetAccountByAccountPk(l.ctx, req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByAccountPk] err:%v", err)
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
