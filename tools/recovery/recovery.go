package recovery

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/tools/recovery/internal/config"
	"github.com/bnb-chain/zkbnb/tools/recovery/internal/svc"
	"github.com/bnb-chain/zkbnb/tree"
)

func RecoveryTreeDB(
	configFile string,
	blockHeight int64,
	serviceName string,
	batchSize int,
) {
	var c config.Config
	conf.MustLoad(configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	// dbinitializer tree database
	treeCtx, err := tree.NewContext(serviceName, c.TreeDB.Driver, true, false, c.TreeDB.RoutinePoolSize, &c.TreeDB.LevelDBOption, &c.TreeDB.RedisDBOption)
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
		c.TreeDB.AssetTreeCacheSize,
	)
	if err != nil {
		logx.Error("InitMerkleTree error:", err)
		return
	}
	// dbinitializer nftTree
	_, err = tree.InitNftTree(
		ctx.NftHistoryModel,
		blockHeight,
		treeCtx)
	if err != nil {
		logx.Errorf("InitNftTree error: %s", err.Error())
		return
	}
}
