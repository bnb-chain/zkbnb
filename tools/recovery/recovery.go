package recovery

import (
	"github.com/bnb-chain/zkbnb/tools/query"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/tools/query/svc"
	"github.com/bnb-chain/zkbnb/tree"
)

func RecoveryTreeDB(
	configFile string,
	blockHeight int64,
	serviceName string,
	batchSize int,
	fromHistory bool,
) {
	configInfo := query.BuildConfig(configFile, serviceName)
	ctx := svc.NewServiceContext(configInfo)
	logx.MustSetup(configInfo.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	// dbinitializer tree database
	treeCtx, err := tree.NewContext(serviceName, configInfo.TreeDB.Driver, true, false, configInfo.TreeDB.RoutinePoolSize, &configInfo.TreeDB.LevelDBOption, &configInfo.TreeDB.RedisDBOption)
	if err != nil {
		logx.Errorf("Init tree database failed: %s", err)
		return
	}

	treeCtx.SetOptions(bsmt.InitializeVersion(0))
	treeCtx.SetBatchReloadSize(batchSize)
	err = tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Errorf("Init tree database failed: %s", err)
		return
	}

	// dbinitializer accountTree and accountStateTrees
	_, _, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		make([]int64, 0),
		blockHeight,
		treeCtx,
		configInfo.TreeDB.AssetTreeCacheSize,
		fromHistory,
	)
	if err != nil {
		logx.Error("InitMerkleTree error:", err)
		return
	}
	logx.Infof("recovery account smt successfully")

	// dbinitializer nftTree
	_, err = tree.InitNftTree(
		ctx.NftModel,
		ctx.NftHistoryModel,
		blockHeight,
		treeCtx, fromHistory)
	if err != nil {
		logx.Errorf("InitNftTree error: %s", err.Error())
		return
	}
	logx.Infof("recovery nft smt successfully")

}
