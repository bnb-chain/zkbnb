package apiserver

import (
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/robfig/cron/v3"
	"net/http"
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

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronJob.AddFunc("@every 1s", func() {
		_, err := ctx.MemCache.SetTxPendingCountKeyPrefix(func() (interface{}, error) {
			txStatuses := []int64{tx.StatusPending}
			return ctx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
		})
		if err != nil {
			logx.Errorf("set tx pending count failed:%s", err.Error())
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		if ctx != nil {
			ctx.Shutdown()
		}
	})

	server := rest.MustNewServer(c.RestConf, rest.WithCors())

	// 全局中间件
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			if request.RequestURI == "/api/v1/sendTx" {
				ctx.SendTxTotalMetrics.Inc()
			}
			next(writer, request)
		}
	})

	handler.RegisterHandlers(server, ctx)

	logx.Infof("apiserver is starting at %s:%d...\n", c.Host, c.Port)
	server.Start()
	return nil
}
