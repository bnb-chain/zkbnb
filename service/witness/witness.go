package witness

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/service/witness/witness"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	w, err := witness.NewWitness(c)
	if err != nil {
		panic(err)
	}
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 2s", func() {
		logx.Info("==========start generate block witness==========")
		err := w.GenerateBlockWitness()
		if err != nil {
			logx.Errorf("failed to generate block witness, %v", err)
		}
		w.RescheduleBlockWitness()
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown witness")
		_ = logx.Close()
		<-cronJob.Stop().Done()
		exit <- struct{}{}
	})

	logx.Info("witness cronjob is starting......")
	select {
	case <-exit:
		break
	}
	return nil
}
