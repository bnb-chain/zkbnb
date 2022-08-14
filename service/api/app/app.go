package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
	"github.com/bnb-chain/zkbas/service/api/app/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	ctx.MemCache.PreloadAccounts()

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()
	handler.RegisterHandlers(server, ctx)
	logx.Infof("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
