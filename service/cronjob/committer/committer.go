package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/svc"
)

var configFile = flag.String("f",
	"./etc/committer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logic.TxsAmountPerBlock = c.KeyPath.KeyTxCounts
	ctx := svc.NewServiceContext(c)
	logx.DisableStat()
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
		if err := logic.CommitterTask(ctx, &lastCommitTimeStamp, accountTree, liquidityTree, nftTree, &accountStateTrees); err != nil {
			cbh, err := ctx.BlockModel.GetCurrentBlockHeight()
			if err != nil {
				logx.Errorf("[GetCurrentBlockHeight] err: %s", err.Error())
				panic("merkle tree re-init.GetCurrentBlockHeight error")
			}
			accountTree, accountStateTrees, err = tree.InitAccountTree(ctx.AccountModel, ctx.AccountHistoryModel, cbh)
			if err != nil {
				logx.Error("[committer] => Re-Init MerkleTree error:", err)
				panic("merkle tree re-init error")
			}
			// init nft tree
			nftTree, err = tree.InitNftTree(ctx.L2NftHistoryModel, cbh)
			if err != nil {
				logx.Error("[committer] => InitMerkleTree error:", err)
				return
			}
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()
	fmt.Printf("Starting committer cronjob ...")
	select {}
}
