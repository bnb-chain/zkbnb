package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
)

var configFile = flag.String("f",
	"./etc/local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	ZecreyRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.ZecreyContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[monitor] fatal error, cannot fetch ZecreyLegendContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZecreyContractAddrSysConfigName)
		panic(err)
	}

	NetworkRpc, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[monitor] fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}

	logx.Infof("[monitor] ChainName: %s, ZecreyRollupAddress: %s, NetworkRpc: %s",
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

	// block monitor
	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start monitor blocks =========================")
		err := logic.MonitorBlocks(
			zecreyRpcCli,
			c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
			ZecreyRollupAddress.Value,
			ctx.L1BlockMonitorModel,
		)
		if err != nil {
			logx.Error("[logic.MonitorBlocks main] unable to run:", err)
		}
		logx.Info("========================= end monitor blocks =========================")
	})
	if err != nil {
		panic(err)
	}

	// mempool monitor
	_, err = cronjob.AddFunc("@every 15s", func() {
		logx.Info("===== start monitor mempool txs")
		err := logic.MonitorMempool(
			ctx,
		)
		if err != nil {
			if err == logic.ErrNotFound {
				logx.Info("[logic.MonitorMempool main] no l2 tx event need to monitor")
			} else {
				logx.Info("[logic.MonitorMempool main] unable to run:", err)
			}
		}
		logx.Info("===== end monitor mempool txs")
	})

	// l2 block monitor
	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start l2 monitor blocks =========================")
		err := logic.MonitorL2BlockEvents(
			zecreyRpcCli,
			c.ChainConfig.PendingBlocksCount,
			ctx.MempoolModel,
			ctx.BlockModel,
			ctx.L1TxSenderModel,
		)
		if err != nil {
			logx.Error("[logic.MonitorBlocks main] unable to run:", err)
		}
		logx.Info("========================= end l2 monitor blocks =========================")
	})
	if err != nil {
		panic(err)
	}

	// governance monitor
	GovernanceContractAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.GovernanceContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[monitor] fatal error, cannot fetch ZecreyLegendContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.GovernanceContractAddrSysConfigName)
		panic(err)
	}

	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start monitor blocks =========================")
		err := logic.MonitorGovernanceContract(
			zecreyRpcCli,
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

	cronjob.Start()

	logx.Info("Starting monitor cronjob ...")
	select {}
}
