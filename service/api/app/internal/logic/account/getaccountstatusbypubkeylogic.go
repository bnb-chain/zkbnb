package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountStatusByPubKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountStatusByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByPubKeyLogic {
	return &GetAccountStatusByPubKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountStatusByPubKeyLogic) GetAccountStatusByPubKey(req *types.ReqGetAccountStatusByPubKey) (resp *types.RespGetAccountStatusByPubKey, err error) {
	// todo: add your logic here and delete this line

	return
}
