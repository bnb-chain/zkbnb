package logic

import (
	"context"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/committerProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/svc"

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

func (l *PingLogic) Ping(in *committerProto.Request) (*committerProto.Response, error) {
	// todo: add your logic here and delete this line

	return &committerProto.Response{}, nil
}
