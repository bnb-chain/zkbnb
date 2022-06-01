package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByAccountIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByAccountIndexLogic {
	return &GetTxsListByAccountIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsListByAccountIndexLogic) GetTxsListByAccountIndex(req *types.ReqGetTxsListByAccountIndex) (resp *types.RespGetTxsListByAccountIndex, err error) {
	// todo: add your logic here and delete this line

	return
}
