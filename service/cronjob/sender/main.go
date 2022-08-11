package main

import (
	"flag"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/svc"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	sender := svc.NewSender(c)
	logx.DisableStat()

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start sender commit task =========================")
		err := sender.CommittedBlocks()
		if err != nil {
			logx.Errorf("[sender.CommittedBlocks] unable to run: %", err)
		} else {
			logx.Info("========================= end sender commit task =========================")
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start sender verify task =========================")
		err = sender.VerifyAndExecuteBlocks()
		if err != nil {
			logx.Errorf("[sender.VerifyAndExecuteBlocks] unable to run: %v", err)
		} else {
			logx.Info("========================= end sender verify task =========================")
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("sender cronjob is starting......")
	select {}
}
