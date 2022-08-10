package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
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
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetMempoolTxs) (*types.RespGetMempoolTxs, error) {
	resp := &types.RespGetMempoolTxs{
		MempoolTxs: make([]*types.Tx, 0),
	}
	count, err := l.svcCtx.MempoolModel.GetMempoolTxsTotalCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	if count == 0 {
		return resp, nil
	}
	mempoolTxs, err := l.svcCtx.MempoolModel.GetMempoolTxsList(int64(req.Offset), int64(req.Limit))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	for _, mempoolTx := range mempoolTxs {
		resp.MempoolTxs = append(resp.MempoolTxs, utils.MempoolTx2Tx(mempoolTx))
	}
	return resp, nil
}
