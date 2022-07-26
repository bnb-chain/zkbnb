package main

import (
	"context"
	"flag"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/svc"
)

var configFile = flag.String("f",
	"./etc/monitor.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	ZkbasRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.ZkbasContractAddrSysConfigName)
	if err != nil {
		logx.Errorf("[main] GetSysconfigByName err: %s", err)
		panic(err)
	}
	NetworkRpc, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[monitor] fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	logx.Infof("[monitor] ChainName: %s, ZkbasRollupAddress: %s, NetworkRpc: %s", c.ChainConfig.ZkbasContractAddrSysConfigName, ZkbasRollupAddress.Value, NetworkRpc.Value)
	zkbasRpcCli, err := _rpc.NewClient(NetworkRpc.Value)
	if err != nil {
		panic(err)
	}
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	if _, err = cronjob.AddFunc("@every 10s", func() {
		logic.MonitorBlocks(zkbasRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount,
			c.ChainConfig.MaxHandledBlocksCount, ZkbasRollupAddress.Value, ctx.L1BlockMonitorModel)
	}); err != nil {
		panic(err)
	}
	if _, err = cronjob.AddFunc("@every 10s", func() {
		logic.MonitorMempool(context.Background(), ctx)
	}); err != nil {
		panic(err)
	}
	if _, err = cronjob.AddFunc("@every 10s", func() {
		logic.MonitorL2BlockEvents(context.Background(), ctx, zkbasRpcCli, c.ChainConfig.PendingBlocksCount,
			ctx.MempoolModel, ctx.BlockModel, ctx.L1TxSenderModel)
	}); err != nil {
		panic(err)
	}
	// governance monitor
	GovernanceContractAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.GovernanceContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[monitor] fatal error, cannot fetch ZkbasContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.GovernanceContractAddrSysConfigName)
		panic(err)
	}

	if _, err = cronjob.AddFunc("@every 10s", func() {
		logic.MonitorGovernanceContract(zkbasRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			GovernanceContractAddress.Value, ctx.L1BlockMonitorModel, ctx.SysConfigModel, ctx.L2AssetInfoModel,
		)
	}); err != nil {
		panic(err)
	}
	cronjob.Start()
	logx.Info("Starting monitor cronjob ...")
	select {}
}
