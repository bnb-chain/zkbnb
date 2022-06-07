package logic

import (
	"context"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/governanceMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *governanceMonitor.Request) (*governanceMonitor.Response, error) {
	// todo: add your logic here and delete this line

	return &governanceMonitor.Response{}, nil
}
