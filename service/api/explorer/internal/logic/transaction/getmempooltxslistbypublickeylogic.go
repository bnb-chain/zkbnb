package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsListByPublicKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMempoolTxsListByPublicKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListByPublicKeyLogic {
	return &GetMempoolTxsListByPublicKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMempoolTxsListByPublicKeyLogic) GetMempoolTxsListByPublicKey(req *types.ReqGetMempoolTxsListByPublicKey) (resp *types.RespGetMempoolTxsListByPublicKey, err error) {
	// todo: add your logic here and delete this line

	return
}
