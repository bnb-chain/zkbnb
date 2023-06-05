package sender

import (
	"github.com/bnb-chain/zkbnb/core/rpc_client"
	"github.com/robfig/cron/v3"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/service/sender/sender"
)

const GracefulShutdownTimeout = 10 * time.Second

func Run(configFile string) error {
	var c config.Config
	if err := config.InitSystemConfiguration(&c, configFile); err != nil {
		logx.Severef("failed to initiate system configuration, %v", err)
		panic("failed to initiate system configuration, err:" + err.Error())
	}

	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	//Initiate the apollo configuration
	config.InitSenderConfiguration(c)
	//Initiate the Prometheus Monitor Facility
	sender.InitPrometheusFacility()

	s := sender.NewSender(c)
	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start commit task =========================")
		if config.GetSenderConfig().DisableCommitBlock {
			logx.Info("disable commit block")
			return
		}
		err := s.CommitBlocks()
		if err != nil {
			logx.Severef("failed to rollup block, %v", err)
		}
	})
	if err != nil {
		logx.Severef("failed to start the commit block task, %v", err)
		panic("failed to start the commit block task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start verify task =========================")
		if config.GetSenderConfig().DisableVerifyBlock {
			logx.Info("disable verify block")
			return
		}
		err = s.VerifyAndExecuteBlocks()
		if err != nil {
			logx.Errorf("failed to send verify transaction, %s", err.Error())
		}
	})
	if err != nil {
		logx.Severef("failed to start the verify and execute block task, %v", err)
		panic("failed to start the verify and execute block task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start update txs task =========================")
		err = s.UpdateSentTxs()
		if err != nil {
			logx.Severef("failed to update update tx status, %v", err)
		}
	})
	if err != nil {
		logx.Severef("failed to start the update send transaction task, %v", err)
		panic("failed to start the update send transaction task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 15s", func() {
		logx.Info("========================= start monitor balance task =========================")
		s.Monitor()
	})
	if err != nil {
		logx.Severef("failed to start the monitor balance task, %v", err)
		panic("failed to start the monitor balance task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 15s", func() {
		logx.Info("========================= start monitor timeout task =========================")
		s.TimeOut()
	})
	if err != nil {
		logx.Severef("failed to start the monitor timeout task, %v", err)
		panic("failed to start the monitor timeout task, err:" + err.Error())
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start rpc health task =========================")
		rpc_client.HealthCheck()
	})
	if err != nil {
		logx.Severef("failed to start rpc health task, %v", err)
		panic("failed to start the rpc health task, err:" + err.Error())
	}

	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown sender......")
		<-cronJob.Stop().Done()
		s.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("sender cronjob is starting......")

	<-exit
	return nil
}
