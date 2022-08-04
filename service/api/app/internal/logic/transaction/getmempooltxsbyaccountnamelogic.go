package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
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
	//TODO: check AccountName
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	mempoolTxDetails, err := l.memPoolTxDetail.GetMemPoolTxDetailByAccountIndex(l.ctx, account.AccountIndex)
	if err != nil {
		logx.Errorf("[GetMemPoolTxDetailByAccountIndex] AccountIndex: %d err: %s", account.AccountIndex, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp := &types.RespGetmempoolTxsByAccountName{
		Txs: make([]*types.Tx, 0),
	}
	for _, d := range mempoolTxDetails {
		tx, err := l.mempool.GetMempoolTxByTxId(l.ctx, d.TxId)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxID] TxId: %d, err: %s", d.TxId, err.Error())
			continue
		}
		resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
