package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsAccountNameRegisteredLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsAccountNameRegisteredLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsAccountNameRegisteredLogic {
	return &IsAccountNameRegisteredLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsAccountNameRegisteredLogic) IsAccountNameRegistered(req *types.ReqIsAccountNameRegistered) (resp *types.RespIsAccountNameRegistered, err error) {
	// todo: add your logic here and delete this line

	return
}
