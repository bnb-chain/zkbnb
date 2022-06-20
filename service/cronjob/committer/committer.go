package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/committer/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/committer/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/committer/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f",
	"./etc/committer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	var (
		accountTree       *tree.Tree
		accountStateTrees []*tree.Tree
		liquidityTree     *tree.Tree
		nftTree           *tree.Tree
	)
	// get latest account
	h, err := ctx.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic(err)
	}
	// init accountTree and accountStateTrees
	accountTree, accountStateTrees, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		h,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}
	// init nft tree
	nftTree, err = tree.InitNftTree(
		ctx.L2NftHistoryModel,
		h,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}

	// init liquidity tree
	liquidityTree, err = tree.InitLiquidityTree(
		ctx.LiquidityHistoryModel,
		h,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}

	/*
		First read the account generalAsset liquidityAsset lockAsset information from the database,
		and then read from the verifier table which layer2 height the information in the database belongs to
	*/
	var lastCommitTimeStamp = time.Now()
	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start committer task =========================")
		err := logic.CommitterTask(
			ctx,
			lastCommitTimeStamp,
			accountTree,
			liquidityTree,
			nftTree,
			&accountStateTrees,
		)
		if err != nil {
			logx.Info("[committer.CommitterTask main] unable to run:", err)

			cbh, err := ctx.BlockModel.GetCurrentBlockHeight()
			if err != nil {
				logx.Error("[committer] => Re-Init MerkleTree GetCurrentBlockHeight error:", err)
				panic("merkle tree re-init.GetCurrentBlockHeight error")
			}

			accountTree, accountStateTrees, err = tree.InitAccountTree(
				ctx.AccountModel,
				ctx.AccountHistoryModel,
				cbh,
			)
			if err != nil {
				logx.Error("[committer] => Re-Init MerkleTree error:", err)
				panic("merkle tree re-init error")
			}
			// init nft tree
			nftTree, err = tree.InitNftTree(
				ctx.L2NftHistoryModel,
				cbh,
			)
			if err != nil {
				logx.Error("[committer] => InitMerkleTree error:", err)
				return
			}
		}
		logx.Info("========================= end committer task =========================")
		// update time stamp
		lastCommitTimeStamp = time.Now()
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	fmt.Printf("Starting committer cronjob ...")
	select {}
}
