package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
	"github.com/bnb-chain/zkbas/service/api/app/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

var (
	CodeVersion   = ""
	GitCommitHash = ""
)

func main() {
	args := os.Args
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-v") {
		fmt.Printf("Git Commit Hash: %s\n", GitCommitHash)
		fmt.Printf("Git Code Version : %s\n", CodeVersion)
		return
	}
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.Severef("[config] err:%v", c)
	logx.DisableStat()
	ctx := svc.NewServiceContext(c)
	ctx.CodeVersion = CodeVersion
	ctx.GitCommitHash = GitCommitHash
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()
	handler.RegisterHandlers(server, ctx)
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
