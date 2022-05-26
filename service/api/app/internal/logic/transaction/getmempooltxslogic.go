package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetMempoolTxs) (resp *types.RespGetMempoolTxs, err error) {
	// todo: add your logic here and delete this line

	return
}
