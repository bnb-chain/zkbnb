package svc

import (
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
