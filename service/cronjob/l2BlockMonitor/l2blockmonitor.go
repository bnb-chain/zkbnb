package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/l2BlockMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/l2BlockMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/l2BlockMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"path/filepath"
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir("./service/cronjob/l2BlockMonitor/etc/local.yaml"))
	if err != nil {
		fmt.Println(err)
	}

	var configFile = flag.String("f", filepath.Join(dir, "local.yaml"), "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

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

	fmt.Printf("Starting l2BlockMonitor cronjob ...")
	select {}
}
