package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/sender"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	sender := sender.NewSender(c)
	logx.DisableStat()

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start sender commit task =========================")
		err := sender.CommitBlocks()
		if err != nil {
			logx.Errorf("[sender.CommitBlocks] unable to run: %", err)
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

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start update sent txs task =========================")
		err = sender.UpdateSentTxs()
		if err != nil {
			logx.Info("update sent txs error, err:", err)
		}
		logx.Info("========================= end update sent txs task =========================")
	})
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("sender cronjob is starting......")
	select {}
}
