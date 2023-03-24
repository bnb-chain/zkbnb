package revertblock

import (
	"context"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/config"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
)

func RevertCommittedBlocks(configFile string, blockHeights []int64) (err error) {
	var c config.Config
	conf.MustLoad(configFile, &c)
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if len(blockHeights) <= 0 {
		logx.Infof("blockHeights is empty")
		return nil
	}
	l1RPCEndpoint, err := ctx.SysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, failed to get network rpc configuration, err:%v, SysConfigName:%s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
		panic("failed to get network rpc configuration, err:" + err.Error() + ", SysConfigName:" +
			c.ChainConfig.NetworkRPCSysConfigName)
	}
	rollupAddress, err := ctx.SysConfigModel.GetSysConfigByName(types.ZkBNBContract)
	if err != nil {
		logx.Severef("fatal error, failed to get zkBNB contract configuration, err:%v, SysConfigName:%s",
			err, types.ZkBNBContract)
		panic("fatal error, failed to get zkBNB contract configuration, err:" + err.Error() + "SysConfigName:" +
			types.ZkBNBContract)
	}
	cli, err := rpc.NewClient(l1RPCEndpoint.Value)
	if err != nil {
		logx.Severef("failed to create client instance, %v", err)
		panic("failed to create client instance, err:" + err.Error())
	}
	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		logx.Severef("failed to get chainId, %v", err)
		panic("failed to get chainId, err:" + err.Error())
	}
	authCliRevertBlock, err := rpc.NewAuthClient(c.ChainConfig.RevertBlockSk, chainId)
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	zkbnbInstance, err := zkbnb.LoadZkBNBInstance(cli, rollupAddress.Value)
	if err != nil {
		logx.Severef("failed to load ZkBNB instance, %v", err)
		panic("failed to load ZkBNB instance, err:" + err.Error())
	}

	storedBlockInfoList := make([]zkbnb.StorageStoredBlockInfo, 0)
	for _, blockHeight := range blockHeights {
		blockInfo, err := ctx.BlockModel.GetBlockByHeight(blockHeight)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
		}
		if blockInfo.BlockStatus != block.StatusCommitted {
			return fmt.Errorf("invalid block status, blockHeight=%d,status=%d", blockHeight, blockInfo.BlockStatus)
		}
		storedBlockInfo := chain.ConstructStoredBlockInfo(blockInfo)
		storedBlockInfoList = append(storedBlockInfoList, storedBlockInfo)
	}

	var gasPrice *big.Int
	if c.ChainConfig.GasPrice > 0 {
		gasPrice = big.NewInt(int64(c.ChainConfig.GasPrice))
	} else {
		gasPrice, err = cli.SuggestGasPrice(context.Background())
		if err != nil {
			logx.Errorf("failed to fetch gas price: %v", err)
			return err
		}
	}
	txHash, err := zkbnb.RevertBlocks(
		cli, authCliRevertBlock,
		zkbnbInstance,
		storedBlockInfoList,
		gasPrice,
		c.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send revertBlocks tx, errL %v:%s", err, txHash)
	}
	logx.Infof("revert block success")
	return nil
}
