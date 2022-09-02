package config

import (
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/tree"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	TreeDB struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
	}
	LogConf logx.LogConf
}
