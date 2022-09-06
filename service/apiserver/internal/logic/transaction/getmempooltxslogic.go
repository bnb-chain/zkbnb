package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetMempoolTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMempoolTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsLogic {
	return &GetMempoolTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetRange) (*types.MempoolTxs, error) {
	total, err := l.svcCtx.MempoolModel.GetMempoolTxsTotalCount()
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	resp := &types.MempoolTxs{
		MempoolTxs: make([]*types.Tx, 0),
		Total:      uint32(total),
	}
	if total == 0 {
		return resp, nil
	}

	mempoolTxs, err := l.svcCtx.MempoolModel.GetMempoolTxsList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, types2.AppErrInternal
	}
	for _, t := range mempoolTxs {
		tx := utils.DbMempooltxTx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.MempoolTxs = append(resp.MempoolTxs, tx)
	}
	return resp, nil
}
