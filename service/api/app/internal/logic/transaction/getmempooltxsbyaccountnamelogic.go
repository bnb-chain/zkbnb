package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetmempoolTxsByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetmempoolTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmempoolTxsByAccountNameLogic {
	return &GetmempoolTxsByAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetmempoolTxsByAccountNameLogic) GetmempoolTxsByAccountName(req *types.ReqGetmempoolTxsByAccountName) (*types.RespGetmempoolTxsByAccountName, error) {
	//TODO: check AccountName
	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	mempoolTxDetails, err := l.svcCtx.MempoolDetailModel.GetMempoolTxDetailsByAccountIndex(account.AccountIndex)
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
		tx, err := l.svcCtx.MempoolModel.GetMempoolTxByTxId(d.TxId)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxID] TxId: %d, err: %s", d.TxId, err.Error())
			continue
		}
		resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
