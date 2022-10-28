package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetPendingTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPendingTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPendingTxsLogic {
	return &GetPendingTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPendingTxsLogic) GetPendingTxs(req *types.ReqGetRange) (*types.Txs, error) {

	txStatuses := []int64{tx.StatusPending}

	total, err := l.svcCtx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	resp := &types.Txs{
		Txs:   make([]*types.Tx, 0),
		Total: uint32(total),
	}
	if total == 0 {
		return resp, nil
	}

	pendingTxs, err := l.svcCtx.TxPoolModel.GetTxs(int64(req.Limit), int64(req.Offset), tx.GetTxWithStatuses(txStatuses))
	if err != nil {
		return nil, types2.AppErrInternal
	}
	for _, pendingTx := range pendingTxs {
		tx := utils.ConvertTx(pendingTx)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		if tx.ToAccountIndex >= 0 {
			tx.ToAccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.ToAccountIndex)
		}
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
