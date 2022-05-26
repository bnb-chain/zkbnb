package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountNameLogic {
	return &GetTxsByAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (resp *types.RespGetTxsByAccountName, err error) {
	// todo: add your logic here and delete this line

	return
}
