package main

import (
	"context"
	"flag"
	"math/big"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	chainBasic "github.com/zecrey-labs/zecrey-eth-rpc/zecreyContract/core/zecrey/basic"
	"github.com/zecrey-labs/zecrey/service/cronjob/sender/internal/config"
	"github.com/zecrey-labs/zecrey/service/cronjob/sender/internal/logic"
	"github.com/zecrey-labs/zecrey/service/cronjob/sender/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f",
	"/Users/gavin/Desktop/zecrey-v2/service/rpc/sender/etc/local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	// srv := server.NewSenderServer(ctx)

	networkEndpointName := c.ChainConfig.NetworkRPCSysConfigName
	networkEndpoint, err := ctx.SysConfigModel.GetSysconfigByName(networkEndpointName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch networkEndpoint from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	ZecreyRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.ZecreyContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch ZecreyRollupAddress from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZecreyContractAddrSysConfigName)
		panic(err)
	}
	mainChainIdConfig, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.MainChainIdSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch mainChainId from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.MainChainIdSysConfigName)
		panic(err)
	}
	mainChainId, err := strconv.ParseInt(mainChainIdConfig.Value, 10, 64)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot parse main chain id: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZecreyContractAddrSysConfigName)
		panic(err)
	}

	cli, err := _rpc.NewClient(networkEndpoint.Value)
	if err != nil {
		panic(err)
	}
	var chainId *big.Int
	if c.ChainConfig.L1ChainId == "" {
		chainId, err = cli.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
	} else {
		var (
			isValid bool
		)
		chainId, isValid = new(big.Int).SetString(c.ChainConfig.L1ChainId, 10)
		if !isValid {
			panic("invalid l1 chain id")
		}
	}

	authCli, err := _rpc.NewAuthClient(cli, c.ChainConfig.Sk, chainId)
	if err != nil {
		panic(err)
	}
	zecreyInstance, err := chainBasic.LoadZecreyInstance(cli, ZecreyRollupAddress.Value)
	if err != nil {
		panic(err)
	}
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	var param *logic.SenderParam = &logic.SenderParam{
		Cli:            cli,
		AuthCli:        authCli,
		ZecreyInstance: zecreyInstance,
		ChainId:        c.ChainConfig.L2ChainId,
		Mode:           logic.StandAlone,
		MaxWaitingTime: c.ChainConfig.MaxWaitingTime * time.Second.Milliseconds(),
		MaxBlockCount:  c.ChainConfig.MaxBlockCount,
		MainChainId:    mainChainId,
		GasLimit:       c.ChainConfig.GasLimit,
		GasPrice:       gasPrice,
		//DebugParams: &utils.DebugOptions{FilePrefix: "/Users/gavin/Desktop"},
	}

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {

		logx.Info("========================= start sender committer task =========================")
		err := logic.SendCommittedBlocks(
			param,
			ctx.L1TxSenderModel, ctx.BlockModel,
		)
		if err != nil {
			logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		}
		logx.Info("========================= end sender committer task =========================")

		logx.Info("========================= start sender verifier task =========================")
		err = logic.SendVerifiedBlocks(param, ctx.L1TxSenderModel, ctx.ProofSenderModel)
		if err != nil {
			logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		}
		logx.Info("========================= end sender verifier task =========================")

		/*
			logx.Info("========================= start sender executor task =========================")
			err = logic.SendExecutedBlocks(
				cli, authCli, zecreyInstance,
				ctx.L1TxSenderModel, ctx.BlockModel, ctx.BlockForProverModel,
				c.ChainConfig.L2ChainId, c.ChainConfig.MaxWaitingTime*time.Second.Milliseconds(), c.ChainConfig.MaxBlockCount, mainChainId,
				gasPrice, c.ChainConfig.GasLimit,
			)
			if err != nil {
				logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
			}
			logx.Info("========================= end sender executor task =========================")

		*/
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("sender cron job is starting......")
	select {}
}
