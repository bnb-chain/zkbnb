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
		logx.Error("new committer failed:", err)
		return err
	}

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown committer......")
		committer.Shutdown()
		_ = logx.Close()
	})

	logx.Info("committer is starting......")
	committer.Run()

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start pending_tx number =========================")
		committer.PendingTxNum()
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()
	return nil
}
