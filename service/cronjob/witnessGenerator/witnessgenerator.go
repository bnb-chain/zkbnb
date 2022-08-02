package main

import (
	"flag"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/service/cronjob/witnessGenerator/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/witnessGenerator/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/witnessGenerator/internal/svc"
)

var configFile = flag.String("f", "./etc/witnessGenerator.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.DisableStat()

	p, err := ctx.ProofSenderModel.GetLatestConfirmedProof()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			logx.Error("[prover] => GetLatestConfirmedProof error:", err)
			return
		} else {
			p = &proofSender.ProofSender{
				BlockNumber: 0,
			}
		}
	}
	var (
		accountTree   *tree.Tree
		assetTrees    []*tree.Tree
		liquidityTree *tree.Tree
		nftTree       *tree.Tree
	)
	// init accountTree and accountStateTrees
	// the init block number use the latest sent block
	accountTree, assetTrees, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		p.BlockNumber,
	)
	// the blockHeight depends on the proof start position
	if err != nil {
		logx.Error("[prover] => InitMerkleTree error:", err)
		return
	}

	liquidityTree, err = tree.InitLiquidityTree(ctx.LiquidityHistoryModel, p.BlockNumber)
	if err != nil {
		logx.Errorf("[prover] InitLiquidityTree error: %s", err.Error())
		return
	}
	nftTree, err = tree.InitNftTree(ctx.NftHistoryModel, p.BlockNumber)
	if err != nil {
		logx.Errorf("[prover] InitNftTree error: %s", err.Error())
		return
	}

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		// cron job for creating cryptoBlock
		logx.Info("==========start generate block witness==========")
		logic.GenerateWitness(
			accountTree,
			&assetTrees,
			liquidityTree,
			nftTree,
			ctx,
			logic.BlockProcessDelta,
		)
		logx.Info("==========end generate block witness==========")
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("witness generator cronjob is starting......")

	select {}
}
