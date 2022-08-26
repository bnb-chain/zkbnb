package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/committer/committer"
)

var configFile = flag.String("f", "/Users/liguo/zkbas-deploy/zkbas/service/committer/etc/config.yaml", "the config file")

func main() {
	flag.Parse()
	var config committer.Config
	conf.MustLoad(*configFile, &config)

	committer, err := committer.NewCommitter(&config)
	if err != nil {
		logx.Error("new committer failed:", err)
		return
	}

	committer.Run()
}
