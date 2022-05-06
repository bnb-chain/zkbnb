package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/l2BlockMonitor"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\SherLzp\\zecrey\\service\\rpc\\l2BlockMonitor\\etc\\l2blockmonitor.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	srv := server.NewL2BlockMonitorServer(ctx)

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	BSCNetworkRpc, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.BSCNetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch BSC NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.BSCNetworkRPCSysConfigName)
		panic(err)
	}

	// load client
	bscCli, err := _rpc.NewClient(BSCNetworkRpc.Value)
	if err != nil {
		panic(err)
	}
	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start monitor blocks =========================")
		err := logic.MonitorL2BlockEvents(
			bscCli,
			c.ChainConfig.BSCPendingBlocksCount,
			ctx.Mempool,
			ctx.AccountAssetModel, ctx.AccountLiquidityModel, ctx.NftModel,
			ctx.AccountAssetHistoryModel, ctx.AccountLiquidityHistoryModel, ctx.NftHistoryModel,
			ctx.Block,
			ctx.L1TxSender,
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
		l2BlockMonitor.RegisterL2BlockMonitorServer(grpcServer, srv)
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
