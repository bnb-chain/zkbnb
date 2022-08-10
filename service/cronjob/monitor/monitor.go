package main

import (
	"flag"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/svc"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	zkbasRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(sysconfigName.ZkbasContract)
	if err != nil {
		logx.Errorf("GetSysconfigByName err: %s", err.Error())
		panic(err)
	}
	networkRpc, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	logx.Infof("ChainName: %s, zkbasRollupAddress: %s, networkRpc: %s", c.ChainConfig.NetworkRPCSysConfigName, zkbasRollupAddress.Value, networkRpc.Value)
	bscRpcCli, err := _rpc.NewClient(networkRpc.Value)
	if err != nil {
		panic(err)
	}
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// block monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorBlocks(bscRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount,
			c.ChainConfig.MaxHandledBlocksCount, zkbasRollupAddress.Value, ctx.L1BlockMonitorModel, ctx.BlockModel, ctx.MempoolModel)
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

	// governance monitor
	governanceContractAddress, err := ctx.SysConfigModel.GetSysconfigByName(sysconfigName.GovernanceContract)
	if err != nil {
		logx.Severef("fatal error, cannot fetch governance contract from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), sysconfigName.GovernanceContract)
		panic(err)
	}

	// governance monitor
	if _, err = cronjob.AddFunc("@every 10s", func() {
		err := logic.MonitorGovernanceContract(bscRpcCli, c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			governanceContractAddress.Value, ctx.L1BlockMonitorModel, ctx.SysConfigModel, ctx.L2AssetInfoModel,
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
