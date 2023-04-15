package monitor

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/monitor/config"
	"github.com/bnb-chain/zkbnb/service/monitor/monitor"
)

const GracefulShutdownTimeout = 10 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	c.Validate()
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	m, err := monitor.NewMonitor(c)
	if err != nil {
		logx.Severe(err)
		panic("failed to start monitor service " + err.Error())
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	// monitor generic blocks
	if _, err := cronJob.AddFunc("@every 10s", func() {
		err := m.MonitorGenericBlocks()
		if err != nil {
			logx.Severef("monitor blocks error, %v", err)
		}
	}); err != nil {
		logx.Severef("failed to start the monitor generic block task, %v", err)
		panic("failed to start the monitor generic block task, err:" + err.Error())
	}

	// monitor priority requests
	if _, err := cronJob.AddFunc("@every 10s", func() {
		err := m.MonitorPriorityRequests()
		if err != nil {
			logx.Severef("monitor priority requests error, %v", err)
		}
	}); err != nil {
		logx.Severef("failed to start monitor priority request task, %v", err)
		panic("failed to start monitor priority request task, err:" + err.Error())
	}

	// monitor governance blocks
	if _, err := cronJob.AddFunc("@every 10s", func() {
		err := m.MonitorGovernanceBlocks()
		if err != nil {
			logx.Severef("monitor governance blocks error, %v", err)
		}

	}); err != nil {
		logx.Severef("failed to start monitor governance block task, %s", err.Error())
		panic("failed to start monitor governance block task, err:" + err.Error())
	}

	// prune historical blocks
	if _, err := cronJob.AddFunc("@every 30s", func() {
		err := m.CleanHistoryBlocks()
		if err != nil {
			logx.Severef("clean history blocks error, %v", err)
		}
	}); err != nil {
		logx.Severef("failed to start clean history blocks, %s", err.Error())
		panic("failed to start clean history blocks, err:" + err.Error())
	}

	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown monitor......")
		<-cronJob.Stop().Done()
		m.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("monitor cronjob is starting......")

	<-exit
	return nil
}
