package prover

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/prover/config"
	"github.com/bnb-chain/zkbnb/service/prover/prover"
)

const GracefulShutdownTimeout = 30 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	p, _ := prover.NewProver(c)
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("start prover job......")
		// cron job for receiving cryptoBlock and handling
		err := p.ProveBlock()
		if err != nil {
			logx.Severef("failed to generate proof, %v", err)
		}
	})
	if err != nil {
		logx.Severe("failed to start prove block task, %v", err)
		panic("failed to start prove block task, err:" + err.Error())
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown prover......")
		<-cronJob.Stop().Done()
		p.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("prover cronjob is starting......")

	<-exit
	return nil
}
