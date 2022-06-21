package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/svc"
	"path/filepath"
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir("./service/cronjob/mempoolMonitor/etc/local.yaml"))
	if err != nil {
		fmt.Println(err)
	}

	var configFile = flag.String("f", filepath.Join(dir, "local.yaml"), "the config file")

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronjob.AddFunc("@every 15s", func() {
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
