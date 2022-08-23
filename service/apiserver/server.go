package main

import (
	"flag"
	"fmt"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/handler"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/server-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
