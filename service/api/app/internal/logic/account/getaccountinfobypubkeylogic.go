package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByPubKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountInfoByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByPubKeyLogic {
	return &GetAccountInfoByPubKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountInfoByPubKeyLogic) GetAccountInfoByPubKey(req *types.ReqGetAccountInfoByPubKey) (resp *types.RespGetAccountInfoByPubKey, err error) {
	// todo: add your logic here and delete this line

	return
}
