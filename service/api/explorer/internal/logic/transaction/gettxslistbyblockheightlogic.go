package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsListByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByBlockHeightLogic {
	return &GetTxsListByBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsListByBlockHeightLogic) GetTxsListByBlockHeight(req *types.ReqGetTxsListByBlockHeight) (resp *types.RespGetTxsListByBlockHeight, err error) {
	// todo: add your logic here and delete this line

	return
}
