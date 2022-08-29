package config

import (
	"github.com/bnb-chain/zkbas/tree"
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	TreeDB struct {
		Driver        tree.Driver
		LevelDBOption tree.LevelDBOption `json:",optional"`
		RedisDBOption tree.RedisDBOption `json:",optional"`
	}
	LogConf logx.LogConf
}
