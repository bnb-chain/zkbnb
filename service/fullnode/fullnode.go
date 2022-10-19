package fullnode

import (
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/fullnode/fullnode"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var config fullnode.Config
	conf.MustLoad(configFile, &config)
	logx.MustSetup(config.LogConf)
	logx.DisableStat()

	node, err := fullnode.NewFullnode(&config)
	if err != nil {
		logx.Error("new fullnode failed:", err)
		return err
	}

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown fullnode......")
		node.Shutdown()
		_ = logx.Close()
	})
	logx.Info("fullnode is starting......")
	node.Run()
	return nil
}
