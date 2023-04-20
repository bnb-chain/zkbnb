package witness

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/service/witness/witness"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var c config.Config
	if err := config.InitSystemConfiguration(&c, configFile); err != nil {
		logx.Severef("failed to initiate system configuration, %v", err)
		panic("failed to initiate system configuration, err:" + err.Error())
	}

	w, err := witness.NewWitness(c)
	if err != nil {
		logx.Severef("failed to create witness instance, %v", err)
		panic("failed to create witness instance, err:" + err.Error())
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 2s", func() {
		logx.Info("==========start generate block witness==========")
		err := w.GenerateBlockWitness()
		if err != nil {
			logx.Severef("failed to generate block witness, %v", err)
			panic("failed to generate block witness, err:" + err.Error())
		}
		w.RescheduleBlockWitness()
	})
	if err != nil {
		logx.Severef("failed to start generate block witness task, %v", err)
		panic("failed to start generate block witness task, err:" + err.Error())
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown witness......")
		<-cronJob.Stop().Done()
		w.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("witness cronjob is starting......")

	<-exit
	return nil
}
