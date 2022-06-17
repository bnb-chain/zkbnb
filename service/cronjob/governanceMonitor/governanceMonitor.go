package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/governanceMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"path/filepath"
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir("./service/cronjob/governanceMonitor/etc/local.yaml"))
	if err != nil {
		fmt.Println(err)
	}

	var configFile = flag.String("f", filepath.Join(dir, "local.yaml"), "the config file")
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	GovernanceContractAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.GovernanceContractAddrSysConfigName)

	if err != nil {
		logx.Severef("[governanceMonitor] fatal error, cannot fetch ZecreyLegendContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.GovernanceContractAddrSysConfigName)
		panic(err)
	}

	NetworkRpc, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[governanceMonitor] fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}

	logx.Infof("[governanceMonitor] ChainName: %s, GovernanceContractAddress: %s, NetworkRpc: %s",
		c.ChainConfig.GovernanceContractAddrSysConfigName,
		GovernanceContractAddress.Value,
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
		err := logic.MonitorGovernanceContract(
			cli,
			c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			GovernanceContractAddress.Value,
			ctx.L1BlockMonitorModel,
			ctx.SysConfigModel,
			ctx.L2AssetInfoModel,
		)
		if err != nil {
			logx.Error("[logic.MonitorGovernanceContract main] unable to run:", err)
		}
		logx.Info("========================= end monitor blocks =========================")
	})
	if err != nil {
		panic(err)
	}
	cronjob.Start()

	fmt.Printf("Starting GovernanceMonitor cronjob ...")
	select{}
}
