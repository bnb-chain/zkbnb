package block

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetCurrentBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentBlockHeightLogic {
	return &GetCurrentBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentBlockHeightLogic) GetCurrentBlockHeight() (resp *types.RespCurrentBlockHeight, err error) {
	resp = &types.RespCurrentBlockHeight{}
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
