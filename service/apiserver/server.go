package apiserver

import (
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/robfig/cron/v3"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"

	_ "github.com/bnb-chain/zkbnb/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"

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
		defer func() {
			if err := recover(); err != nil {
				logx.Severef("set tx pending count failed:%v", err)
			}
		}()
		_, err := ctx.MemCache.SetTxPendingCountKeyPrefix(func() (interface{}, error) {
			txStatuses := []int64{tx.StatusPending}
			return ctx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
		})
		if err != nil {
			logx.Severef("set tx pending count failed:%s", err.Error())
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 500s", func() {
		defer func() {
			if err := recover(); err != nil {
				logx.Severef("get layer 2 basic info failed:%v", err)
			}
		}()
		l := info.NewGetLayer2BasicInfoLogic(nil, ctx)
		_, err := l.GetLayer2BasicInfo(false)
		if err != nil {
			logx.Severef("get layer 2 basic info failed:%s", err.Error())
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

	// Start the swagger server in background
	go startSwaggerServer()

	logx.Infof("apiserver is starting at %s:%d...\n", c.Host, c.Port)
	server.Start()
	return nil
}

func startSwaggerServer() {

	logx.Infof("swagger server is starting at port:%d", 8866)
	engine := gin.Default()
	engine.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))

	err := engine.Run(":8866")
	if err != nil {
		logx.Errorf("swagger server fails to start! err:%s", err)
	}
}
