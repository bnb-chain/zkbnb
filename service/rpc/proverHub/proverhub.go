package main

import (
	"flag"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/model/proofSender"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/proverHubProto"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f",
	"./etc/proverhub.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	srv := server.NewProverHubRPCServer(ctx)

	p, err := ctx.ProofSenderModel.GetLatestConfirmedProof()
	if err != nil {
		if err != proofSender.ErrNotFound {
			logx.Error("[prover] => GetLatestConfirmedProof error:", err)
			return
		} else {
			p = &proofSender.ProofSender{
				BlockNumber: 0,
			}
		}
	}
	var (
		accountTree   *tree.Tree
		assetTrees    []*tree.Tree
		liquidityTree *tree.Tree
		nftTree       *tree.Tree
	)
	// init accountTree and accountStateTrees
	// the init block number use the latest sent block
	accountTree, assetTrees, err = tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		p.BlockNumber,
	)
	// the blockHeight depends on the proof start position
	if err != nil {
		logx.Error("[prover] => InitMerkleTree error:", err)
		return
	}

	liquidityTree, err = tree.InitLiquidityTree(ctx.LiquidityHistoryModel, 1)
	if err != nil {
		logx.Errorf("[prover] InitLiquidityTree error: %s", err)
		return
	}
	nftTree, err = tree.InitNftTree(ctx.NftHistoryModel, 1)
	if err != nil {
		logx.Errorf("[prover] InitNftTree error: %s", err)
		return
	}
	// TODO
	logic.VerifyingKeyPath = c.KeyPath.VerifyingKeyPath

	err = logic.InitUnprovedList(
		accountTree,
		&assetTrees,
		liquidityTree,
		nftTree,
		ctx,
		p.BlockNumber)

	if err != nil {
		logx.Error("[prover] => InitUnprovedList error:", err)
		return
	}

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 10s", func() {
		// cron job for creating cryptoBlock
		logx.Info("==========start handle crypto block==========")
		for _, v := range logic.UnProvedCryptoBlocks {
			logx.Infof("BlockNumber: %v, Status: %d", v.BlockInfo.BlockNumber, v.Status)
		}
		logic.HandleCryptoBlock(
			accountTree,
			&assetTrees,
			liquidityTree,
			nftTree,
			ctx,
			10, // TODO
		)
		logx.Info("==========end handle crypto block==========")
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proverHubProto.RegisterProverHubRPCServer(grpcServer, srv)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
