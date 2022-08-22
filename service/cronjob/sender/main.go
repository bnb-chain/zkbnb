package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/sender"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	s := sender.NewSender(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start s commit task =========================")
		err := s.CommitBlocks()
		if err != nil {
			logx.Errorf("[s.CommitBlocks] unable to run: %", err)
		} else {
			logx.Info("========================= end s commit task =========================")
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start s verify task =========================")
		err = s.VerifyAndExecuteBlocks()
		if err != nil {
			logx.Errorf("[s.VerifyAndExecuteBlocks] unable to run: %v", err)
		} else {
			logx.Info("========================= end s verify task =========================")
		}
	})

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start update sent txs task =========================")
		err = s.UpdateSentTxs()
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

	logx.Info("s cronjob is starting......")
	select {}
}
