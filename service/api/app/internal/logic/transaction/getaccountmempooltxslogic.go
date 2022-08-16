package transaction

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountMempoolTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountMempoolTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountMempoolTxsLogic {
	return &GetAccountMempoolTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountMempoolTxsLogic) GetAccountMempoolTxs(req *types.ReqGetAccountMempoolTxs) (resp *types.MempoolTxs, err error) {
	accountIndex := int64(0)
	switch req.By {
	case "account_index":
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for account_index")
		}
	case "account_name":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case "account_pk":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be account_index|account_name|account_pk")
	}

	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	mempoolTxs, err := l.svcCtx.MempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	resp = &types.MempoolTxs{
		Total:      uint32(len(mempoolTxs)),
		MempoolTxs: make([]*types.Tx, 0),
	}
	for _, t := range mempoolTxs {
		tx := utils.DbMempoolTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.MempoolTxs = append(resp.MempoolTxs, tx)
	}
	return resp, nil
}
