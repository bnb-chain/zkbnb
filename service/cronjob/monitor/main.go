package main

import (
	"flag"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/monitor/config"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/monitor"
)

var configFile = flag.String("f",
	"./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	m := monitor.NewMonitor(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// m generic blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorGenericBlocks()
		if err != nil {
			logx.Errorf("monitor blocks error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// m priority requests
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorPriorityRequests()
		if err != nil {
			logx.Errorf("monitor priority requests error, err=%s", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// m governance blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorGovernanceBlocks()
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
