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

const (
	queryByAccountIndex = "account_index"
	queryByAccountName  = "account_name"
	queryByAccountPk    = "account_pk"
)

type GetAccountTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountTxsLogic {
	return &GetAccountTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountTxsLogic) GetAccountTxs(req *types.ReqGetAccountTxs) (resp *types.Txs, err error) {
	resp = &types.Txs{
		Txs: make([]*types.Tx, 0),
	}

	accountIndex := int64(0)
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil || accountIndex < 0 {
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

	total, err := l.svcCtx.TxModel.GetTxsCountByAccountIndex(accountIndex)
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	resp.Total = uint32(total)
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	txs, err := l.svcCtx.TxModel.GetTxsByAccountIndex(accountIndex, int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, types2.AppErrInternal
	}

	for _, t := range txs {
		tx := utils.DbtxTx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
