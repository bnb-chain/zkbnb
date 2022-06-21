package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/config"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

var configFile = flag.String("f", "etc/explorer-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
