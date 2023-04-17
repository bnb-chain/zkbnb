package estimategas

import (
	"context"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/service/sender/sender"
	"github.com/bnb-chain/zkbnb/tools/estimategas/internal/config"
	"github.com/bnb-chain/zkbnb/tools/estimategas/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
)

func EstimateGas(configFile string, fromHeight int64, toHeight int64, maxBlockCount int64) (err error) {
	var c config.Config
	conf.MustLoad(configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if fromHeight == 0 || toHeight == 0 || maxBlockCount == 0 || maxBlockCount > (toHeight-fromHeight+1) {
		return fmt.Errorf("input parameter error")
	}

	l1RPCEndpoint, err := ctx.SysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		return fmt.Errorf("fatal error, failed to get network rpc configuration, err:%v, SysConfigName:%s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
	}
	rollupAddress, err := ctx.SysConfigModel.GetSysConfigByName(types.ZkBNBContract)
	if err != nil {
		return fmt.Errorf("fatal error, failed to get zkBNB contract configuration, err:%v, SysConfigName:%s",
			err, types.ZkBNBContract)
	}
	cli, err := rpc.NewClient(l1RPCEndpoint.Value)
	if err != nil {
		return fmt.Errorf("failed to create client instance, %v", err)
	}
	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chainId, %v", err)
	}
	authCliCommitBlock, err := rpc.NewAuthClient(c.ChainConfig.CommitBlockSk, chainId)
	if err != nil {
		return fmt.Errorf("failed to create authClient error, %v", err)
	}

	zkBNBClient, err := zkbnb.NewZkBNBClient(cli, rollupAddress.Value)
	if err != nil {
		return fmt.Errorf("failed to initiate ZkBNBClient instance, %v", err)
	}
	zkBNBClient.CommitConstructor = authCliCommitBlock

	blocks, err := ctx.CompressedBlockModel.GetCompressedBlocksBetween(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	blockCount := int64(0)
	for blockCount <= maxBlockCount {
		totalRealTxCount := CalculateRealBlockTxCount(blocks[0:blockCount])
		totalTxCount := CalculateTotalTxCount(blocks[0:blockCount])

		pendingCommitBlocks, err := sender.ConvertBlocksForCommitToCommitBlockInfos(blocks, ctx.TxModel)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		lastStoredBlockInfo := sender.DefaultBlockHeader()
		lastHandledBlockInfo, err := ctx.BlockModel.GetBlockByHeight(blocks[0].BlockHeight)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		lastStoredBlockInfo = chain.ConstructStoredBlockInfo(lastHandledBlockInfo)
		//nonce, err :=cli.GetPendingNonce(s.GetVerifyAddress().Hex())
		//if err != nil {
		//	return fmt.Errorf("failed to get nonce for verify block, errL %v", err)
		//}
		var gasPrice *big.Int
		gasPrice, err = cli.SuggestGasPrice(context.Background())
		if err != nil {
			return fmt.Errorf("failed to fetch gas price: %v", err)
		}
		estimatedFee, err := zkBNBClient.EstimateCommitGasWithNonce(lastStoredBlockInfo, pendingCommitBlocks, gasPrice, c.ChainConfig.GasLimit, 0)
		if err != nil {
			return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
		}
		logx.Infof("estimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d,totalTxCount=%d", estimatedFee, totalRealTxCount, totalTxCount)
		blockCount++
	}
	logx.Infof("estimate gas success")
	return nil
}

func CalculateRealBlockTxCount(blocks []*compressedblock.CompressedBlock) uint64 {
	var totalTxCount uint16 = 0
	if len(blocks) > 0 {
		for _, b := range blocks {
			totalTxCount = totalTxCount + b.RealBlockSize
		}
	}
	return uint64(totalTxCount)
}

func CalculateTotalTxCount(blocks []*compressedblock.CompressedBlock) uint64 {
	var totalTxCount uint16 = 0
	if len(blocks) > 0 {
		for _, b := range blocks {
			totalTxCount = totalTxCount + b.BlockSize
		}
		return uint64(totalTxCount)
	}
	return 0
}
