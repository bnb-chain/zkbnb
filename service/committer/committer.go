package committer

import (
	"github.com/bnb-chain/zkbnb/types"
	"github.com/robfig/cron/v3"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/committer/committer"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var c committer.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	committer, err := committer.NewCommitter(&c)
	if err != nil {
		logx.Severef("failed to create committer instance, %v", err)
		// If the rollback fails, wait 1 minute
		time.Sleep(1 * time.Minute)
		panic("failed to create committer instance, err:" + err.Error())
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		committer.PendingTxNum()
	})
	if err != nil {
		logx.Severef("failed to add PendingTxNum cron job, %v", err)
		panic("failed to add PendingTxNum cron job, err:" + err.Error())
	}

	if _, err := cronJob.AddFunc("@every 300s", func() {
		committer.CompensatePendingPoolTx()
	}); err != nil {
		logx.Severef("failed to start the compensate pending pool transaction task, %v", err)
		panic("failed to start the compensate pending pool transaction task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= update NFT index =========================")
		err = committer.SyncNftIndexServer()
		if err != nil {
			if err != types.DbErrNotFound {
				logx.Severef("failed to update NFT index, %v", err)
			}
		}
	})
	if err != nil {
		logx.Severe("failed to start the sync nft index server task, %v", err)
		panic("failed to start the sync nft index server task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= send message to ipns =========================")
		err = committer.SendIpfsServer()
		if err != nil {
			if err != types.DbErrNotFound {
				logx.Severef("failed to send message to ipns, %v", err)
			}
		}
	})
	if err != nil {
		logx.Severef("failed to start send ipfs server task, %v", err)
		panic("failed to start send ipfs server task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 6h", func() {
		logx.Info("========================= send message to refresh ipns =========================")
		err = committer.RefreshServer()
		if err != nil {
			if err != types.DbErrNotFound {
				logx.Severef("failed to send message to refresh ipns, %v", err)
			}
		}
	})
	if err != nil {
		logx.Severef("failed to start the refresh server task, %v", err)
		panic("failed to start the refresh server task, err:" + err.Error())
	}

	cronJob.Start()

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown committer......")
		committer.Shutdown()
		_ = logx.Close()
	})

	logx.Info("committer is starting......")
	err = committer.Run()
	if err != nil {
		logx.Severef("failed to committer.run, %v", err)
		panic("failed to committer.run, err:" + err.Error())
	}
	return nil
}
