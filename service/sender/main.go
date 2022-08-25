package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbas/service/sender/config"
	"github.com/bnb-chain/zkbas/service/sender/sender"
)

var configFile = flag.String("f", "./etc/config.yaml", "the path of config file")

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
		logx.Info("========================= start commit task =========================")
		err := s.CommitBlocks()
		if err != nil {
			logx.Errorf("failed to rollup blocks: %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start verify task =========================")
		err = s.VerifyAndExecuteBlocks()
		if err != nil {
			logx.Errorf("failed to send verify transaction %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start update txs task =========================")
		err = s.UpdateSentTxs()
		if err != nil {
			logx.Errorf("failed to update update tx status, err: %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	cronJob.Start()

	logx.Info("cronjob is starting......")
	select {}
}
