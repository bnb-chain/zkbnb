package committer

import (
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
		logx.Severef("new committer failed: %v", err)
		return err
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		committer.PendingTxNum()
	})
	if err != nil {
		logx.Severef("add PendingTxNum cron job failed: %v", err)
		return err
	}

	if _, err := cronJob.AddFunc("@every 300s", func() {
		committer.CompensatePendingPoolTx()
	}); err != nil {
		logx.Severe(err)
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= update NFT index =========================")
		err = committer.SyncNftIndexServer()
		if err != nil {
			logx.Severef("failed to update NFT index, %v", err)
		}
	})
	if err != nil {
		logx.Severe(err)
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= send message to ipns =========================")
		err = committer.SendIpfsServer()
		if err != nil {
			logx.Severef("failed to send message to ipns, %v", err)
		}
	})
	if err != nil {
		logx.Severe(err)
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 6h", func() {
		logx.Info("========================= send message to refresh ipns =========================")
		err = committer.RefreshServer()
		if err != nil {
			logx.Severef("failed to send message to refresh ipns, %v", err)
		}
	})
	if err != nil {
		logx.Severe(err)
		panic(err)
	}

	cronJob.Start()

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown committer......")
		committer.Shutdown()
		_ = logx.Close()
	})

	logx.Info("committer is starting......")
	committer.Run()
	return nil
}
