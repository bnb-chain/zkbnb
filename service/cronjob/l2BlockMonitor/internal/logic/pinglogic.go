package logic

import (
	"context"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/l2BlockMonitor/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/l2BlockMonitor/l2BlockMonitor"
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

func (l *PingLogic) Ping(in *l2BlockMonitor.Request) (*l2BlockMonitor.Response, error) {
	// todo: add your logic here and delete this line

	return &l2BlockMonitor.Response{}, nil
}
