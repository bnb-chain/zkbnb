package main

import (
	"flag"
	"fmt"
	blockmonitor "github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/blockMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/blockMonitor/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\SherLzp\\zecrey-legend\\service\\rpc\\blockMonitor\\etc\\local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	srv := server.NewBlockMonitorServer(ctx)

	ZecreyRollupAddress, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.ZecreyContractAddrSysConfigName)

	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch ZecreyLegendContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZecreyContractAddrSysConfigName)
		panic(err)
	}

	NetworkRpc, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}

	logx.Infof("[blockMonitor] ChainName: %s, ZecreyRollupAddress: %s, NetworkRpc: %s",
		c.ChainConfig.ZecreyContractAddrSysConfigName,
		ZecreyRollupAddress.Value,
		NetworkRpc.Value)

	// load client
	cli, err := _rpc.NewClient(NetworkRpc.Value)
	if err != nil {
		panic(err)
	}

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start monitor blocks =========================")
		err := logic.MonitorBlocks(
			cli,
			c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			ZecreyRollupAddress.Value,
			ctx.L1BlockMonitor,
		)
		if err != nil {
			logx.Error("[logic.MonitorBlocks main] unable to run:", err)
		}
		logx.Info("========================= end monitor blocks =========================")
	})
	if err != nil {
		panic(err)
	}
	cronjob.Start()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		blockmonitor.RegisterBlockMonitorServer(grpcServer, srv)
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
