package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAssetIdAndPubKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountInfoByAssetIdAndPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAssetIdAndPubKeyLogic {
	return &GetAccountInfoByAssetIdAndPubKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountInfoByAssetIdAndPubKeyLogic) GetAccountInfoByAssetIdAndPubKey(req *types.ReqGetAccountInfoByAssetIdAndPubKey) (resp *types.RespGetAccountInfoByAssetIdAndPubKey, err error) {
	// todo: add your logic here and delete this line

	return
}
