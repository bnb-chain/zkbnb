package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetsListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsListLogic {
	return &GetAssetsListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetsListLogic) GetAssetsList(req *types.ReqGetAssetsList) (resp *types.RespGetAssetsList, err error) {
	// todo: add your logic here and delete this line

	return
}
