package main

import (
	"flag"

	"github.com/bnb-chain/zkbas/service/witness/config"
	"github.com/bnb-chain/zkbas/service/witness/witness"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

var configFile = flag.String("f", "./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	w := witness.NewWitness(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronJob.AddFunc("@every 2s", func() {
		logx.Info("==========start generate block witness==========")
		err := w.GenerateBlockWitness()
		if err != nil {
			logx.Errorf("==========failed to generate block witness %v==========", err)
		} else {
			logx.Info("==========success generate block witness==========")
		}
		w.RescheduleBlockWitness()
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("witness cronjob is starting......")
	select {}
}
