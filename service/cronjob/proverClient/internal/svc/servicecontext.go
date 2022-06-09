package svc

import (
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/proverClient/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/proverhubrpc"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	ProverHubRPC proverhubrpc.ProverHubRPC
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		ProverHubRPC: proverhubrpc.NewProverHubRPC(zrpc.MustNewClient(c.ProverHubRPC)),
	}
}
