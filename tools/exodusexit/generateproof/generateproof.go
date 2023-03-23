package generateproof

import (
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/config"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/generateproof"
	"github.com/goccy/go-json"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string, address string, token string, nftIndexListStr string, proofFolder string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	if address != "" {
		c.Address = address
	}
	if token != "" {
		c.Token = token
	}
	if proofFolder != "" {
		c.ProofFolder = proofFolder
	}
	if nftIndexListStr != "" {
		var nftIndexList []int64
		err := json.Unmarshal([]byte(nftIndexListStr), &nftIndexList)
		if err != nil {
			return nil
		}
		c.NftIndexList = nftIndexList
	}

	m, err := generateproof.NewMonitor(&c)
	if err != nil {
		logx.Severe(err)
		panic(err)
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
		logx.Severe(err)
		panic(err)
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown generateproof......")
		<-cronJob.Stop().Done()
		m.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})
	exodusExit, err := generateproof.NewExodusExit(&c)
	if err != nil {
		return err
	}
	err = exodusExit.Run()
	if err != nil {
		return err
	}
	logx.Info("generateproof cronjob is starting......")

	<-exit
	return nil
}
