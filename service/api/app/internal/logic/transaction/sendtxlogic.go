package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.RespSendTx, err error) {
	// todo: add your logic here and delete this line

	return
}
