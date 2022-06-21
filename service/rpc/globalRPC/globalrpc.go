package main

import (
	"flag"
	"fmt"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/config"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/server"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"path/filepath"
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir("service/rpc/globalRPC/etc/globalrpc.yaml"))
	if err != nil {
		fmt.Println(err)
	}

	var configFile = flag.String("f", filepath.Join(dir, "globalrpc.yaml"), "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	svr := server.NewGlobalRPCServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		globalRPCProto.RegisterGlobalRPCServer(grpcServer, svr)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
