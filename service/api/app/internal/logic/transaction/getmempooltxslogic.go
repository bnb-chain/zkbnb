package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetAll) (*types.MempoolTxs, error) {
	total, err := l.svcCtx.MempoolModel.GetMempoolTxsTotalCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
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
		return nil, errorcode.AppErrInternal
	}
	for _, t := range mempoolTxs {
		tx := utils.DbMempoolTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.MempoolTxs = append(resp.MempoolTxs, tx)
	}
	return resp, nil
}
