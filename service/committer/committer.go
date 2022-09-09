package committer

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"time"

	"github.com/bnb-chain/zkbnb/service/committer/committer"
)

const GracefulShutdownTimeout = 2 * time.Second

func Run(configFile string) error {
	var config committer.Config
	conf.MustLoad(configFile, &config)
	logx.MustSetup(config.LogConf)
	logx.DisableStat()

	committer, err := committer.NewCommitter(&config)
	if err != nil {
		logx.Error("new committer failed:", err)
		return err
	}

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown committer......")
		_ = logx.Close()
		committer.Shutdown()
	})

	logx.Info("committer is starting......")
	committer.Run()
	return nil
}
