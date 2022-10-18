package transaction

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetAccountPendingTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountPendingTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountPendingTxsLogic {
	return &GetAccountPendingTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountPendingTxsLogic) GetAccountPendingTxs(req *types.ReqGetAccountPendingTxs) (resp *types.Txs, err error) {
	resp = &types.Txs{
		Txs: make([]*types.Tx, 0),
	}

	accountIndex := int64(0)
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil || accountIndex < 0 {
			return nil, types2.AppErrInvalidAccountIndex
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

	poolTxs, err := l.svcCtx.TxPoolModel.GetPendingTxsByAccountIndex(accountIndex)
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	resp.Total = uint32(len(poolTxs))
	for _, poolTx := range poolTxs {
		tx := utils.ConvertTx(poolTx)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		if tx.ToAccountIndex >= 0 {
			tx.ToAccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.ToAccountIndex)
		}
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
