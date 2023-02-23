package exodusexit

import (
	"github.com/bnb-chain/zkbnb/service/exodusexit/exodusexit"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/exodusexit/config"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	m, err := exodusexit.NewMonitor(c)
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	// monitor generic blocks
	if _, err := cronJob.AddFunc("@every 10s", func() {
		err := m.MonitorGenericBlocks()
		if err != nil {
			logx.Severef("monitor blocks error, %v", err)
		}
	}); err != nil {
		logx.Severe(err)
		panic(err)
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown exodusexit......")
		<-cronJob.Stop().Done()
		m.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("exodusexit cronjob is starting......")

	<-exit
	return nil
}
