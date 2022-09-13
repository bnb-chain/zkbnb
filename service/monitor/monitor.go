package monitor

import (
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/monitor/config"
	"github.com/bnb-chain/zkbnb/service/monitor/monitor"
)

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	m := monitor.NewMonitor(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// monitor generic blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorGenericBlocks()
		if err != nil {
			logx.Errorf("monitor blocks error, %v", err)
		}
	}); err != nil {
		panic(err)
	}

	// monitor priority requests
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorPriorityRequests()
		if err != nil {
			logx.Errorf("monitor priority requests error, %v", err)
		}
	}); err != nil {
		panic(err)
	}

	// monitor governance blocks
	if _, err := cronjob.AddFunc("@every 10s", func() {
		err := m.MonitorGovernanceBlocks()
		if err != nil {
			logx.Errorf("monitor governance blocks error, %v", err)
		}

	}); err != nil {
		panic(err)
	}
	cronjob.Start()
	logx.Info("Starting monitor cronjob ...")
	select {}
}
