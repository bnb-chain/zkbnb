package block

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetCurrentHeightLogic) GetCurrentHeight() (resp *types.RespCurrentHeight, err error) {
	resp = &types.RespCurrentHeight{}
	height, err := l.svcCtx.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return resp, nil
		}
		return nil, errorcode.AppErrInternal
	}
	resp.Height = height
	return resp, nil
}
