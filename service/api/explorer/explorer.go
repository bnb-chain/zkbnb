package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/config"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

func main() {
	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir("./service/api/explorer/etc/explorer-api.yaml"))
	if err != nil {
		fmt.Println(err)
	}
	var configFile = flag.String("f", filepath.Join(dir, "explorer-api.yaml"), "the config file")

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
