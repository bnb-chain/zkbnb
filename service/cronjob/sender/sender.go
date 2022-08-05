package main

import (
	"context"
	"math/big"
	"time"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/internal/svc"
)

func main() {
	configFile := util.ReadConfigFileFlag()
	var c config.Config
	conf.MustLoad(configFile, &c)

	ctx := svc.NewServiceContext(c)
	// srv := server.NewSenderServer(ctx)
	logx.DisableStat()
	networkEndpointName := c.ChainConfig.NetworkRPCSysConfigName
	networkEndpoint, err := ctx.SysConfigModel.GetSysconfigByName(networkEndpointName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch networkEndpoint from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	ZkbasRollupAddress, err := ctx.SysConfigModel.GetSysconfigByName(c.ChainConfig.ZkbasContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch ZkbasRollupAddress from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZkbasContractAddrSysConfigName)
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
	zkbasInstance, err := zkbas.LoadZkbasInstance(cli, ZkbasRollupAddress.Value)
	if err != nil {
		panic(err)
	}
	gasPrice, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	var param = &logic.SenderParam{
		Cli:            cli,
		AuthCli:        authCli,
		ZkbasInstance:  zkbasInstance,
		MaxWaitingTime: c.ChainConfig.MaxWaitingTime * time.Second.Milliseconds(),
		MaxBlocksCount: c.ChainConfig.MaxBlockCount,
		GasPrice:       gasPrice,
		GasLimit:       c.ChainConfig.GasLimit,
	}

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start sender committer task =========================")
		err := logic.SendCommittedBlocks(
			param,
			ctx.L1TxSenderModel,
			ctx.BlockModel,
			ctx.BlockForCommitModel,
		)
		if err != nil {
			logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		}
		logx.Info("========================= end sender committer task =========================")
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start sender verifier task =========================")
		err = logic.SendVerifiedAndExecutedBlocks(param, ctx.L1TxSenderModel, ctx.BlockModel, ctx.ProofSenderModel)
		if err != nil {
			logx.Info("[sender.SendCommittedBlocks main] unable to run:", err)
		}
		logx.Info("========================= end sender verifier task =========================")
	})
	if err != nil {
		panic(err)
	}

	cronJob.Start()

	logx.Info("sender cronjob is starting......")
	select {}
}
