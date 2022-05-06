package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/internal/svc"
	mempoolmonitor "github.com/zecrey-labs/zecrey-legend/service/rpc/mempoolMonitor/mempoolMonitor"

	"google.golang.org/grpc"

	"github.com/robfig/cron/v3"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\SherLzp\\zecrey\\service\\rpc\\mempoolMonitor\\etc\\mempoolMonitor.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	srv := server.NewMempoolMonitorServer(ctx)

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronjob.AddFunc("@every 15s", func() {
		logx.Info("===== start monitor mempool txs")
		err := logic.MonitorMempool(
			ctx.L2TxEventMonitorModel, ctx.L2assetInfoModel,
			ctx.AccountModel, ctx.AccountHistoryModel,
			ctx.AccountAssetModel,
			// todo fix ctx param
			ctx, context.Background(),
		)
		if err != nil {
			if err == l2TxEventMonitor.ErrNotFound {
				logx.Info("[mempoolMonitor.MonitorMempool main] no l2 tx event need to monitor")
			} else {
				logx.Info("[mempoolMonitor.MonitorMempool main] unable to run:", err)
			}
		}
		logx.Info("===== end monitor mempool txs")
	})
	if err != nil {
		panic(err)
	}
	cronjob.Start()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		mempoolmonitor.RegisterMempoolMonitorServer(grpcServer, srv)
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
