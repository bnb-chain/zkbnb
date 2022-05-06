package logic

import (
	"context"
	blockmonitor "github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/blockMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/internal/svc"

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

func (l *PingLogic) Ping(in *blockmonitor.Request) (*blockmonitor.Response, error) {
	// todo: add your logic here and delete this line

	return &blockmonitor.Response{}, nil
}
