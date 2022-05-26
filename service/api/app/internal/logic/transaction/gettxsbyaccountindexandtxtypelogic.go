package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexAndTxTypeLogic {
	return &GetTxsByAccountIndexAndTxTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (resp *types.RespGetTxsByAccountIndexAndTxType, err error) {
	// todo: add your logic here and delete this line

	return
}
