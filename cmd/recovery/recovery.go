package main

import (
	"flag"
	"fmt"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbas/cmd/recovery/internal/config"
	"github.com/bnb-chain/zkbas/cmd/recovery/internal/svc"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/treedb"
)

var (
	configFile  = flag.String("f", "./etc/recovery.yaml", "the config file")
	blockHeight = flag.Int64("height", 0, "block height")
	serviceName = flag.String("service", "", "service name(committer, witness)")
	batchSize   = flag.Int("batch", 1000, "batch size")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	if *blockHeight < 0 {
		fmt.Println("-height must be greater than 0")
		flag.Usage()
		return
	}

	if *batchSize <= 0 {
		fmt.Println("-batch must be greater than 0")
		flag.Usage()
		return
	}

	if *serviceName == "" {
		fmt.Println("-service must be set")
		flag.Usage()
		return
	}

	// init tree database
	treeCtx := &treedb.Context{
		Name:          *serviceName,
		Driver:        c.TreeDB.Driver,
		LevelDBOption: &c.TreeDB.LevelDBOption,
		RedisDBOption: &c.TreeDB.RedisDBOption,
		Reload:        true,
	}
	treeCtx.SetOptions(bsmt.InitializeVersion(bsmt.Version(*blockHeight) - 1))
	treeCtx.SetBatchReloadSize(*batchSize)
	err := treedb.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Errorf("Init tree database failed: %s", err)
		return
	}

	// init accountTree and accountStateTrees
	_, _, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		*blockHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("InitMerkleTree error:", err)
		return
	}
	// init liquidityTree
	_, err = tree.InitLiquidityTree(
		ctx.LiquidityHistoryModel,
		*blockHeight,
		treeCtx)
	if err != nil {
		logx.Errorf("InitLiquidityTree error: %s", err.Error())
		return
	}
	// init nftTree
	_, err = tree.InitNftTree(
		ctx.NftHistoryModel,
		*blockHeight,
		treeCtx)
	if err != nil {
		logx.Errorf("InitNftTree error: %s", err.Error())
		return
	}
}
