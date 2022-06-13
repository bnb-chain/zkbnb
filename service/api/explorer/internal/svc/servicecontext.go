package svc

import (
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
