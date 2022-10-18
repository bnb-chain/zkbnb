package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsLogic {
	return &GetTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsLogic) GetTxs(req *types.ReqGetRange) (resp *types.Txs, err error) {
	total, err := l.svcCtx.MemCache.GetTxTotalCountWithFallback(func() (interface{}, error) {
		return l.svcCtx.TxModel.GetTxsTotalCount()
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	resp = &types.Txs{
		Total: uint32(total),
		Txs:   make([]*types.Tx, 0, req.Limit),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	txs, err := l.svcCtx.TxModel.GetTxs(int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, types2.AppErrInternal
	}
	for _, dbTx := range txs {
		tx := utils.ConvertTx(dbTx)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		if tx.ToAccountIndex >= 0 {
			tx.ToAccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.ToAccountIndex)
		}
		resp.Txs = append(resp.Txs, tx)
	}

	return resp, nil
}
