package main

import (
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
		logx.Errorf("GetSysconfigByName err: %s", err.Error())
		panic(err)
	}
	NetworkRpc, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	logx.Infof("ChainName: %s, ZkbasRollupAddress: %s, NetworkRpc: %s", c.ChainConfig.ZkbasContractAddrSysConfigName, ZkbasRollupAddress.Value, NetworkRpc.Value)
	zkbasRpcCli, err := _rpc.NewClient(NetworkRpc.Value)
	if err != nil {
		panic(err)
	}
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// block monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorBlocks(zkbasRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount,
			c.ChainConfig.MaxHandledBlocksCount, ZkbasRollupAddress.Value, ctx.L1BlockMonitorModel)
		if err != nil {
			logx.Errorf("monitor blocks error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// mempool monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorMempool(ctx)
		if err != nil {
			logx.Errorf("monitor mempool error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// l2 block monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorL2BlockEvents(zkbasRpcCli, c.ChainConfig.PendingBlocksCount, ctx.MempoolModel, ctx.BlockModel, ctx.L1TxSenderModel)
		if err != nil {
			logx.Errorf("monitor l2 block events error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}
	// governance monitor
	GovernanceContractAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.GovernanceContractAddrSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch ZkbasContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.GovernanceContractAddrSysConfigName)
		panic(err)
	}

	// governance monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorGovernanceContract(zkbasRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			GovernanceContractAddress.Value, ctx.L1BlockMonitorModel, ctx.SysConfigModel, ctx.L2AssetInfoModel,
		)
		if err != nil {
			logx.Errorf("monitor governance contracts events error, err=%s", err.Error())
		}

	}); err != nil {
		panic(err)
	}
	cronjob.Start()
	logx.Info("Starting monitor cronjob ...")
	select {}
}
