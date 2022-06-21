package svc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/bnb-chain/zkbas/service/cronjob/proverClient/internal/config"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/proverhubrpc"
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
