package estimategas

import (
	"context"
	"encoding/json"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb-eth-rpc/utils"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/service/sender/sender"
	"github.com/bnb-chain/zkbnb/tools/estimategas/internal/config"
	"github.com/bnb-chain/zkbnb/tools/estimategas/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
	"strings"
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

	cli, err := rpc.NewClient(strings.Split(l1RPCEndpoint.Value, ",")[0])
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
	zkBNBClient.VerifyConstructor = authCliCommitBlock

	err = EstimateCommitBlockGas(ctx, cli, zkBNBClient, fromHeight, toHeight, maxBlockCount)
	if err != nil {
		logx.Errorf("failed to EstimateCommitBlockGas, %v", err)
	}

	err = EstimateVerifyBlockGas(ctx, cli, zkBNBClient, fromHeight, toHeight, maxBlockCount)
	if err != nil {
		logx.Errorf("failed to EstimateVerifyBlockGas, %v", err)
	}

	logx.Infof("estimate gas success")
	return nil
}

func Send(configFile string, fromHeight int64, toHeight int64, sendFlag int64) (err error) {
	var c config.Config
	conf.MustLoad(configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if fromHeight == 0 || toHeight == 0 {
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
	cli, err := rpc.NewClient(strings.Split(l1RPCEndpoint.Value, ",")[0])
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
	zkBNBClient.VerifyConstructor = authCliCommitBlock

	if sendFlag == 1 {
		err = CommitBlocks(ctx, cli, zkBNBClient, fromHeight, toHeight)
		if err != nil {
			logx.Errorf("failed to EstimateCommitBlockGas, %v", err)
		}
	}

	if sendFlag == 2 {
		err = VerifyBlocks(ctx, cli, zkBNBClient, fromHeight, toHeight)
		if err != nil {
			logx.Errorf("failed to EstimateVerifyBlockGas, %v", err)
		}
	}

	logx.Infof("estimate gas success")
	return nil
}

func EstimateCommitBlockGas(ctx *svc.ServiceContext, cli *rpc.ProviderClient, zkBNBClient *zkbnb.ZkBNBClient, fromHeight int64, toHeight int64, maxBlockCount int64) error {
	blocks, err := ctx.CompressedBlockModel.GetCompressedBlocksBetween(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	blockCount := int64(1)
	for blockCount <= maxBlockCount {
		totalRealTxCount := CalculateRealBlockTxCount(blocks[0:blockCount])
		totalTxCount := CalculateTotalTxCount(blocks[0:blockCount])

		pendingCommitBlocks, err := sender.ConvertBlocksForCommitToCommitBlockInfos(blocks[0:blockCount], ctx.TxModel)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		lastHandledBlockInfo, err := ctx.BlockModel.GetBlockByHeight(blocks[0].BlockHeight - 1)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		lastStoredBlockInfo := chain.ConstructStoredBlockInfo(lastHandledBlockInfo)

		var gasPrice *big.Int
		gasPrice, err = cli.SuggestGasPrice(context.Background())
		if err != nil {
			return fmt.Errorf("failed to fetch gas price: %v", err)
		}
		nonce, err := GetNonce(ctx, cli)
		if err != nil {
			return err
		}
		estimatedFee, err := zkBNBClient.EstimateCommitGasWithNonce(lastStoredBlockInfo, pendingCommitBlocks, gasPrice, 0, nonce)
		if err != nil {
			return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
		}
		logx.Infof("EstimateCommitGas,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d", blockCount, estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount)
		blockCount++
	}
	return nil
}

func CommitBlocks(ctx *svc.ServiceContext, cli *rpc.ProviderClient, zkBNBClient *zkbnb.ZkBNBClient, fromHeight int64, toHeight int64) error {
	blocks, err := ctx.CompressedBlockModel.GetCompressedBlocksBetween(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	totalRealTxCount := CalculateRealBlockTxCount(blocks)
	totalTxCount := CalculateTotalTxCount(blocks)

	pendingCommitBlocks, err := sender.ConvertBlocksForCommitToCommitBlockInfos(blocks, ctx.TxModel)
	if err != nil {
		return fmt.Errorf("failed to get block info, err: %v", err)
	}
	lastHandledBlockInfo, err := ctx.BlockModel.GetBlockByHeight(blocks[0].BlockHeight - 1)
	if err != nil {
		return fmt.Errorf("failed to get block info, err: %v", err)
	}
	lastStoredBlockInfo := chain.ConstructStoredBlockInfo(lastHandledBlockInfo)

	var gasPrice *big.Int
	gasPrice, err = cli.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch gas price: %v", err)
	}
	nonce, err := GetNonce(ctx, cli)
	if err != nil {
		return err
	}
	estimatedFee, err := zkBNBClient.EstimateCommitGasWithNonce(lastStoredBlockInfo, pendingCommitBlocks, gasPrice, 0, nonce)
	if err != nil {
		return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
	}
	logx.Infof("EstimateCommitGas,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d", len(blocks), estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount)

	tx, err := zkBNBClient.CommitBlocksWithNonce(lastStoredBlockInfo, pendingCommitBlocks, gasPrice, estimatedFee, nonce)
	if err != nil {
		return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
	}
	logx.Infof("CommitBlocksWithNonce,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d,hash=%s", len(blocks), estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount, tx)

	return nil
}

func EstimateVerifyBlockGas(ctx *svc.ServiceContext, cli *rpc.ProviderClient, zkBNBClient *zkbnb.ZkBNBClient, fromHeight int64, toHeight int64, maxBlockCount int64) error {
	compressedBlocks, err := ctx.CompressedBlockModel.GetCompressedBlocksBetween(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}

	blocks, err := ctx.BlockModel.GetPendingBlocksBetweenWithoutTx(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("unable to get blocks, err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	blockCount := int64(1)
	for blockCount <= maxBlockCount {
		totalRealTxCount := CalculateRealBlockTxCount(compressedBlocks[0:blockCount])
		totalTxCount := CalculateTotalTxCount(compressedBlocks[0:blockCount])

		pendingVerifyAndExecuteBlocks, err := sender.ConvertBlocksToVerifyAndExecuteBlockInfos(blocks[0:blockCount])
		if err != nil {
			return fmt.Errorf("unable to convert blocks to commit block infos: %v", err)
		}

		blockProofs, err := ctx.ProofModel.GetProofsBetween(blocks[0].BlockHeight, blocks[blockCount-1].BlockHeight)
		if err != nil {
			if err == types.DbErrNotFound {
				return nil
			}
			return fmt.Errorf("unable to get proofs, err: %v", err)
		}
		if len(blockProofs) != len(blocks[0:blockCount]) {
			return types.AppErrRelatedProofsNotReady
		}
		// add sanity check
		for i := range blockProofs {
			if blockProofs[i].BlockNumber != blocks[i].BlockHeight {
				return types.AppErrProofNumberNotMatch
			}
		}
		var proofs []*big.Int
		for _, bProof := range blockProofs {
			var proofInfo *prove.FormattedProof
			err = json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo)
			if err != nil {
				return err
			}
			proofs = append(proofs, proofInfo.A[:]...)
			proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
			proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
			proofs = append(proofs, proofInfo.C[:]...)
		}

		var gasPrice *big.Int
		gasPrice, err = cli.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		nonce, err := GetNonce(ctx, cli)
		if err != nil {
			return err
		}
		estimatedFee, err := zkBNBClient.EstimateVerifyAndExecuteWithNonce(pendingVerifyAndExecuteBlocks, proofs, gasPrice, 0, nonce)
		if err != nil {
			return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
		}
		logx.Infof("EstimateVerifyAndExecuteGas,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d,", blockCount, estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount)
		blockCount++
	}
	return nil
}

func VerifyBlocks(ctx *svc.ServiceContext, cli *rpc.ProviderClient, zkBNBClient *zkbnb.ZkBNBClient, fromHeight int64, toHeight int64) error {
	compressedBlocks, err := ctx.CompressedBlockModel.GetCompressedBlocksBetween(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}

	blocks, err := ctx.BlockModel.GetPendingBlocksBetweenWithoutTx(fromHeight,
		toHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("unable to get blocks, err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	totalRealTxCount := CalculateRealBlockTxCount(compressedBlocks)
	totalTxCount := CalculateTotalTxCount(compressedBlocks)

	pendingVerifyAndExecuteBlocks, err := sender.ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
	if err != nil {
		return fmt.Errorf("unable to convert blocks to commit block infos: %v", err)
	}

	blockProofs, err := ctx.ProofModel.GetProofsBetween(fromHeight, toHeight)
	if err != nil {
		if err == types.DbErrNotFound {
			return nil
		}
		return fmt.Errorf("unable to get proofs, err: %v", err)
	}
	if len(blockProofs) != len(blocks) {
		return types.AppErrRelatedProofsNotReady
	}
	// add sanity check
	for i := range blockProofs {
		if blockProofs[i].BlockNumber != blocks[i].BlockHeight {
			return types.AppErrProofNumberNotMatch
		}
	}
	var proofs []*big.Int
	for _, bProof := range blockProofs {
		var proofInfo *prove.FormattedProof
		err = json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo)
		if err != nil {
			return err
		}
		proofs = append(proofs, proofInfo.A[:]...)
		proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
		proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
		proofs = append(proofs, proofInfo.C[:]...)
	}

	var gasPrice *big.Int
	gasPrice, err = cli.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	nonce, err := GetNonce(ctx, cli)
	if err != nil {
		return err
	}
	estimatedFee, err := zkBNBClient.EstimateVerifyAndExecuteWithNonce(pendingVerifyAndExecuteBlocks, proofs, gasPrice, 0, nonce)
	if err != nil {
		return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
	}
	logx.Infof("EstimateVerifyAndExecuteGas,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d,", len(blocks), estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount)

	tx, err := zkBNBClient.VerifyAndExecuteBlocksWithNonce(pendingVerifyAndExecuteBlocks, proofs, gasPrice, estimatedFee, nonce)
	if err != nil {
		return fmt.Errorf("abandon send block to l1, EstimateGas operation get some error:%s", err.Error())
	}
	logx.Infof("EstimateVerifyAndExecuteGas,blockCount=%d,totalEstimatedFee=%d,averageEstimatedFee=%d,totalTxCount=%d,totalRealTxCount=%d,hash=%s", len(blocks), estimatedFee, estimatedFee/totalTxCount, totalTxCount, totalRealTxCount, tx)

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

func GetNonce(ctx *svc.ServiceContext, cli *rpc.ProviderClient) (uint64, error) {
	privateKey, err := utils.DecodePrivateKey(ctx.Config.ChainConfig.CommitBlockSk)
	if err != nil {
		return 0, err
	}
	// get public key
	publicKey := privateKey.PublicKey
	address := crypto.PubkeyToAddress(publicKey)
	nonce, err := cli.GetPendingNonce(address.Hex())
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce for send block, errL %v", err)
	}
	return nonce, nil
}
