package main

import (
	"flag"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/witness/config"
	"github.com/bnb-chain/zkbas/service/cronjob/witness/witness"
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
		// cron job for creating cryptoBlock
		logx.Info("==========start generate block witness==========")
		w.GenerateBlockWitness()
		logx.Info("==========end generate block witness==========")
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("witness cronjob is starting......")

	select {}
}
