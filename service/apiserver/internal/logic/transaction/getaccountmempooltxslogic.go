package transaction

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
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
	resp = &types.MempoolTxs{
		MempoolTxs: make([]*types.Tx, 0),
	}

	accountIndex := int64(0)
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, types2.AppErrInvalidParam.RefineError("invalid value for account_index")
		}
	case queryByAccountName:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case queryByAccountPk:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be account_index|account_name|account_pk")
	}

	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, types2.AppErrInternal
	}

	mempoolTxs, err := l.svcCtx.MempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	resp.Total = uint32(len(mempoolTxs))
	for _, t := range mempoolTxs {
		tx := utils.DbMempooltxTx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.MempoolTxs = append(resp.MempoolTxs, tx)
	}
	return resp, nil
}
