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
	if !utils.ValidateAccountName(req.AccountName) {
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	mempoolTxDetails, err := l.svcCtx.MempoolDetailModel.GetMempoolTxDetailsByAccountIndex(account.AccountIndex)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	resp := &types.RespGetmempoolTxsByAccountName{
		MempoolTxs: make([]*types.Tx, 0),
	}
	for _, d := range mempoolTxDetails {
		tx, err := l.svcCtx.MempoolModel.GetMempoolTxByTxId(d.TxId)
		if err != nil {
			continue
		}
		resp.MempoolTxs = append(resp.MempoolTxs, utils.MempoolTx2Tx(tx))
	}
	return resp, nil
}
