package transaction

import (
	"context"

	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExecutedTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetExecutedTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExecutedTxsLogic {
	return &GetExecutedTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetExecutedTxsLogic) GetExecutedTxs(req *types.ReqGetRangeWithFromHash) (*types.Txs, error) {

	options := []tx.GetTxOptionFunc{
		tx.GetTxWithStatuses([]int64{tx.StatusExecuted}),
		tx.GetTxWithDeleted(),
	}
	if len(req.FromHash) > 0 {
		options = append(options, tx.GetTxWithFromHash(req.FromHash))
	}

	total, err := l.svcCtx.TxPoolModel.GetTxsTotalCount(options...)
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

	pendingTxs, err := l.svcCtx.TxPoolModel.GetTxs(int64(req.Limit), int64(req.Offset), options...)
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
