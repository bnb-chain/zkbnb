package main

import (
	"flag"
	"fmt"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/zecrey-labs/zecrey-crypto/zecrey-legend/circuit/bn254/block"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/proverClient/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/proverClient/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/proverClient/internal/svc"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f",
	"./etc/proverClient.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	// srv := server.NewProverClientPingServer(ctx)
	logx.DisableStat()

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	logic.KeyTxCounts = c.KeyPath.KeyTxCounts
	logic.ProvingKeys = make([]groth16.ProvingKey, len(logic.KeyTxCounts))
	logic.VerifyingKeys = make([]groth16.VerifyingKey, len(logic.KeyTxCounts))
	logic.R1cs = make([]frontend.CompiledConstraintSystem, len(logic.KeyTxCounts))
	var err error
	for i := 0; i < len(logic.KeyTxCounts); i++ {
		var circuit block.BlockConstraints
		circuit.TxsCount = logic.KeyTxCounts[i]
		circuit.Txs = make([]block.TxConstraints, circuit.TxsCount)
		for i := 0; i < circuit.TxsCount; i++ {
			circuit.Txs[i] = block.GetZeroTxConstraint()
		}
		fmt.Printf("start compile block size %d circuit\n", circuit.TxsCount)
		logic.R1cs[i], err = frontend.Compile(ecc.BN254, r1cs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
		if err != nil {
			panic("r1cs init error")
		}
		fmt.Println("circuit constraints:", logic.R1cs[i].GetNbConstraints())
		fmt.Println("finish compile circuit")
		// read proving and verifying keys
		logic.ProvingKeys[i], err = util.LoadProvingKey(c.KeyPath.ProvingKeyPath[i])
		if err != nil {
			panic("provingKey loading error")
		}
		logic.VerifyingKeys[i], err = util.LoadVerifyingKey(c.KeyPath.VerifyingKeyPath[i])
		if err != nil {
			panic("verifyingKey loading error")
		}
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		// cron job for receiving cryptoBlock and handling
		err = logic.ProveBlock(ctx)
		if err != nil {
			logx.Error("Prove Error: ", err.Error())
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	logx.Info("proverClient cronjob is starting......")
	select {}
}
