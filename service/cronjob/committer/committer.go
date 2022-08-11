package main

import (
	"flag"
	"fmt"
	"time"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/pkg/treedb"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/svc"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logic.TxsAmountPerBlock = c.KeyPath.KeyTxCounts
	ctx := svc.NewServiceContext(c)
	logx.DisableStat()
	var (
		accountTree       bsmt.SparseMerkleTree
		accountStateTrees []bsmt.SparseMerkleTree
		liquidityTree     bsmt.SparseMerkleTree
		nftTree           bsmt.SparseMerkleTree
	)
	// get latest account
	h, err := ctx.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic(err)
	}
	latestVerifiedBlockNr, err := ctx.BlockModel.GetLatestVerifiedBlockHeight()
	if err != nil {
		panic(err)
	}
	// init tree database
	treeCtx := &treedb.Context{
		Name:          "committer",
		Driver:        c.TreeDB.Driver,
		LevelDBOption: &c.TreeDB.LevelDBOption,
		RedisDBOption: &c.TreeDB.RedisDBOption,
	}
	err = treedb.SetupTreeDB(treeCtx)
	if err != nil {
		panic(err)
	}
	// init accountTree and accountStateTrees
	accountTree, accountStateTrees, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		h,
		treeCtx,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}
	// init nft tree
	nftTree, err = tree.InitNftTree(
		ctx.L2NftHistoryModel,
		h,
		treeCtx,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}

	// init liquidity tree
	liquidityTree, err = tree.InitLiquidityTree(
		ctx.LiquidityHistoryModel,
		h,
		treeCtx,
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
			&lastCommitTimeStamp,
			treeCtx,
			accountTree,
			liquidityTree,
			nftTree,
			&accountStateTrees,
			uint64(latestVerifiedBlockNr),
		)
		if err != nil {
			logx.Info("[committer.CommitterTask main] unable to run:", err)

			accountTree.Reset()
			for _, assetTree := range accountStateTrees {
				assetTree.Reset()
			}
			nftTree.Reset()
			liquidityTree.Reset()
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()
	fmt.Printf("Starting committer cronjob ...")
	select {}
}
