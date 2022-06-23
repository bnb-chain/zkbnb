package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetmempoolTxsByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetmempoolTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmempoolTxsByAccountNameLogic {
	return &GetmempoolTxsByAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetmempoolTxsByAccountNameLogic) GetmempoolTxsByAccountName(req *types.ReqGetmempoolTxsByAccountName) (resp *types.RespGetmempoolTxsByAccountName, err error) {
	// todo: add your logic here and delete this line

	return
}
