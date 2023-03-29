package info

import (
	"context"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProtocolRateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProtocolRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProtocolRateLogic {
	return &GetProtocolRateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProtocolRateLogic) GetProtocolRate(fromCache bool) (resp *types.ProtocolRate, err error) {
	protocolRate, err := l.svcCtx.MemCache.GetSysConfigWithFallback(types2.ProtocolRate, fromCache, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(types2.ProtocolRate)
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	return &types.ProtocolRate{
		ProtocolRate: protocolRate.Value,
	}, nil
}
