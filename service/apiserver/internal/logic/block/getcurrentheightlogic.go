package block

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

type GetCurrentHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentHeightLogic {
	return &GetCurrentHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentHeightLogic) GetCurrentHeight() (resp *types.CurrentHeight, err error) {
	resp = &types.CurrentHeight{}
	height, err := l.svcCtx.BlockModel.GetCurrentHeight()
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, types2.AppErrInternal
	}
	resp.Height = height
	return resp, nil
}
