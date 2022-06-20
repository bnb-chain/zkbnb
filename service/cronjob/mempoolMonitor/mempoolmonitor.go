package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/mempoolMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/mempoolMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/mempoolMonitor/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f",
	"./etc/local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronjob.AddFunc("@every 15s", func() {
		logx.Info("===== start monitor mempool txs")
		err := logic.MonitorMempool(
			ctx,
		)
		if err != nil {
			if err == l2TxEventMonitor.ErrNotFound {
				logx.Info("[mempoolMonitor.MonitorMempool main] no l2 tx event need to monitor")
			} else {
				logx.Info("[mempoolMonitor.MonitorMempool main] unable to run:", err)
			}
		}
		logx.Info("===== end monitor mempool txs")
	})
	if err != nil {
		panic(err)
	}
	cronjob.Start()

	fmt.Printf("Starting MempoolMonitor cronjob ...")
	select {}
}
