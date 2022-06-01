package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMempoolTxsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListLogic {
	return &GetMempoolTxsListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMempoolTxsListLogic) GetMempoolTxsList(req *types.ReqGetMempoolTxsList) (resp *types.RespGetMempoolTxsList, err error) {
	// todo: add your logic here and delete this line

	return
}
