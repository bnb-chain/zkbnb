package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"path/filepath"
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir("./service/cronjob/blockMonitor/etc/local.yaml"))
	if err != nil {
		fmt.Println(err)
	}

	var configFile = flag.String("f", filepath.Join(dir, "local.yaml"), "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

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
	zecreyRpcCli, err := _rpc.NewClient(NetworkRpc.Value)
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
			zecreyRpcCli,
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


	fmt.Printf("Starting BlockMonitor cronjob ...")
	select {}
}
