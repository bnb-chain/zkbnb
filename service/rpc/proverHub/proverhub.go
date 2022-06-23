package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"path/filepath"

	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/config"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/logic"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/server"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/svc"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/proverHubProto"
)

func main() {
	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir("./service/rpc/proverHub/etc/proverhub.yaml"))
	if err != nil {
		fmt.Println(err)
	}
	var configFile = flag.String("f", filepath.Join(dir, "proverhub.yaml"), "the config file")

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

	liquidityTree, err = tree.InitLiquidityTree(ctx.LiquidityHistoryModel, p.BlockNumber)
	if err != nil {
		logx.Errorf("[prover] InitLiquidityTree error: %s", err)
		return
	}
	nftTree, err = tree.InitNftTree(ctx.NftHistoryModel, p.BlockNumber)
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
