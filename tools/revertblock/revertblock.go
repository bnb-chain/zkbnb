package revertblock

import (
	"context"
	"fmt"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/config"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
	"sort"
	"strings"
	"time"
)

func RevertCommittedBlocks(configFile string, height int64) (err error) {
	c := config.Config{}
	if err := config.InitSystemConfiguration(&c, configFile); err != nil {
		logx.Severef("failed to initiate system configuration, %v", err)
		panic("failed to initiate system configuration, err:" + err.Error())
	}

	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if height == 0 {
		return fmt.Errorf("height can not be 0")
	}
	startHeight := height
	lastHandledTx, err := ctx.L1RollupTxModel.GetLatestHandledTx(l1rolluptx.TxTypeCommit)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("lastHandledTx is nil %v", err)
	}
	endHeight := int64(0)
	if lastHandledTx != nil {
		endHeight = lastHandledTx.L2BlockHeight
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

	zkBNBClient, err := zkbnb.NewZkBNBClient(cli, rollupAddress.Value)
	if err != nil {
		return fmt.Errorf("failed to initiate ZkBNBClient instance, %v", err)
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		logx.Severef("failed to load KMS client config, %v", err)
		panic("failed to load KMS client config, err:" + err.Error())
	}
	kmsClient := kms.NewFromConfig(cfg)
	commitKeyId := c.KMSConfig.CommitKeyId
	if len(commitKeyId) > 0 {
		commitKmsKeyClient, err := rpc.NewKMSKeyClient(kmsClient, c.KMSConfig.CommitKeyId, chainId)
		if err != nil {
			logx.Severef("fatal error, failed to initiate commit kmsKeyClient instance, err:%v", err)
			panic("fatal error, failed to initiate commit kmsKeyClient instance, err:" + err.Error())
		}
		zkBNBClient.RevertConstructor = commitKmsKeyClient
	} else {
		commitBlockSk := c.AuthConfig.CommitBlockSk
		commitAuthClient, err := rpc.NewAuthClient(commitBlockSk, chainId)
		if err != nil {
			logx.Severef("fatal error, failed to initiate commit authClient instance, err:%v", err)
			panic("fatal error, failed to initiate commit authClient instance, err:" + err.Error())
		}
		zkBNBClient.RevertConstructor = commitAuthClient
	}
	storedBlockInfoList := make([]zkbnb.StorageStoredBlockInfo, 0)
	for height <= endHeight {
		blockInfo, err := ctx.BlockModel.GetBlockByHeight(height)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		if blockInfo.BlockStatus != block.StatusCommitted {
			return fmt.Errorf("invalid block status, blockHeight=%d,status=%d", height, blockInfo.BlockStatus)
		}
		storedBlockInfo := chain.ConstructStoredBlockInfo(blockInfo)
		storedBlockInfoList = append(storedBlockInfoList, storedBlockInfo)
		height++
	}

	sort.Slice(storedBlockInfoList, func(i, j int) bool {
		return storedBlockInfoList[i].BlockNumber > storedBlockInfoList[j].BlockNumber
	})

	var gasPrice *big.Int
	if c.ChainConfig.GasPrice > 0 {
		gasPrice = big.NewInt(int64(c.ChainConfig.GasPrice))
	} else {
		gasPrice, err = cli.SuggestGasPrice(context.Background())
		if err != nil {
			return fmt.Errorf("failed to fetch gas price: %v", err)
		}
	}
	txHash, err := zkBNBClient.RevertBlocks(
		storedBlockInfoList,
		gasPrice,
		c.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send revertBlocks tx, errL %v:%s", err, txHash)
	}
	logx.Infof("send revert block success,tx hash=%s,startHeight=%d ~ endHeight=%d", txHash, startHeight, endHeight)
	err = checkRevertBlock(cli, c, txHash)
	if err != nil {
		logx.Severe(err)
		return nil
	}
	logx.Infof("revert block success,tx hash=%s,startHeight=%d ~ endHeight=%d", txHash, startHeight, endHeight)
	return nil
}

func checkRevertBlock(cli *rpc.ProviderClient, c config.Config, txHash string) error {
	startDate := time.Now()
	for {
		receipt, err := cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("query transaction receipt %s failed, err: %v", txHash, err)
			if time.Now().After(startDate.Add(time.Duration(c.ChainConfig.MaxWaitingTime) * time.Second)) {
				return fmt.Errorf("failed to revert block, tx_hash=%s,error=%s", txHash, err)
			}
			continue
		}
		if receipt.Status == 0 {
			return fmt.Errorf("failed to revert block, tx_hash=%s,receipt.Status=0", txHash)
		}
		latestL1Height, err := cli.GetHeight()
		if err != nil {
			return fmt.Errorf("failed to get l1 block height, err: %v", err)
		}
		if latestL1Height < receipt.BlockNumber.Uint64()+c.ChainConfig.ConfirmBlocksCount {
			continue
		} else {
			return nil
		}
	}
}
