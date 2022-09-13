package apiserver

import (
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/handler"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

const GracefulShutdownTimeout = 5 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	ctx := svc.NewServiceContext(c)
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		if ctx != nil {
			ctx.Shutdown()
		}
	})

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	handler.RegisterHandlers(server, ctx)

	logx.Infof("apiserver is starting at %s:%d...\n", c.Host, c.Port)
	server.Start()
	return nil
}
