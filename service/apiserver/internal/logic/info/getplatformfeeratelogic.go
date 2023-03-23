package info

import (
	"context"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformFeeRateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlatformFeeRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformFeeRateLogic {
	return &GetPlatformFeeRateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlatformFeeRateLogic) GetPlatformFeeRate(fromCache bool) (resp *types.PlatformFeeRate, err error) {
	platformFeeRate, err := l.svcCtx.MemCache.GetSysConfigWithFallback("PlatformFeeRate", fromCache, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName("PlatformFeeRate")
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	return &types.PlatformFeeRate{
		PlatformFeeRate: platformFeeRate.Value,
	}, nil
}
