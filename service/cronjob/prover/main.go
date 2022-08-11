package main

import (
	"flag"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/cronjob/prover/config"
	"github.com/bnb-chain/zkbas/service/cronjob/prover/prover"
)

var configFile = flag.String("f", "./etc/config.yaml", "the path of config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	prover := prover.NewProver(c)
	logx.DisableStat()

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("start prover job......")
		// cron job for receiving cryptoBlock and handling
		err := prover.ProveBlock()
		if err != nil {
			logx.Error("Prove Error: ", err.Error())
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("prover cronjob is starting......")
	select {}
}
