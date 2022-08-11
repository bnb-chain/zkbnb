package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/monitor/config"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	monitor := monitor.NewMonitor(c)

	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// monitor generic blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := monitor.MonitorGenericBlocks()
		if err != nil {
			logx.Errorf("monitor blocks error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// monitor priority requests
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := monitor.MonitorPriorityRequests()
		if err != nil {
			logx.Errorf("monitor priority requests error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// monitor governance blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := monitor.MonitorGovernanceBlocks()
		if err != nil {
			logx.Errorf("monitor governance blocks error, err=%s", err.Error())
		}

	}); err != nil {
		panic(err)
	}
	cronjob.Start()
	logx.Info("Starting monitor cronjob ...")
	select {}
}
