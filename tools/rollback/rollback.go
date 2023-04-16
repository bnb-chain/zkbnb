package rollback

import (
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/tools/revertblock"
	"github.com/bnb-chain/zkbnb/tools/rollback/internal/config"
	"github.com/bnb-chain/zkbnb/tools/rollback/internal/svc"
	"github.com/bnb-chain/zkbnb/tools/rollbackwitnesssmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

//If the smt tree data is incorrect, automatic rollback cannot be used

func RollbackAll(configFile string, height int64) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	logx.Infof("revert CommittedBlocks,start height=%d", height)
	err := revertblock.RevertCommittedBlocks(configFile, height)
	if err != nil {
		return err
	}

	logx.Infof("delete L1RollupTx,start height=%d", height)
	err = ctx.L1RollupTxModel.DeleteGreaterOrEqualToHeight(height, l1rolluptx.TxTypeCommit)
	if err != nil {
		return err
	}

	logx.Infof("update block status to StatusPending,start height=%d", height)
	err = ctx.BlockModel.UpdateGreaterOrEqualHeight(height, block.StatusPending)
	if err != nil {
		return err
	}

	logx.Infof("delete proof,start height=%d", height)
	err = ctx.ProofModel.DeleteGreaterOrEqualToHeight(height)
	if err != nil {
		logx.Severe(err)
		return err
	}

	logx.Infof("roll back witness smt tree,start height=%d", height)
	err = rollbackwitnesssmt.RollbackWitnessSmt(configFile, height)
	if err != nil {
		return err
	}

	logx.Infof("delete block witness,start height=%d", height)
	err = ctx.BlockWitnessModel.DeleteGreaterOrEqualToHeight(height)
	if err != nil {
		return err
	}

	logx.Infof("update block status to StatusProposing,start height=%d", height)
	err = ctx.BlockModel.UpdateGreaterOrEqualHeight(height, block.StatusProposing)
	if err != nil {
		return err
	}
	logx.Infof("rollback success,start height=%d", height)

	return nil
}
