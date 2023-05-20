package revertblock

import (
	"context"
	"fmt"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/abicoder"
	"github.com/bnb-chain/zkbnb/common/chain"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/tools/desertexit/desertexit"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
	"sort"
	"strings"
	"time"
)

func RevertCommittedBlocks(configFile string, height int64, byBlock bool) (err error) {
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

	if !c.EnableRollback {
		return fmt.Errorf("rollback switch not turned on")
	}

	if height == 0 {
		return fmt.Errorf("height can not be 0")
	}
	startHeight := height

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

	totalBlocksCommitted, err := getTotalBlocksCommitted(cli, rollupAddress.Value)
	if err != nil {
		return err
	}
	logx.Infof("the last committed height is %d", totalBlocksCommitted)

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
	if byBlock == true {
		storedBlockInfoList, err = buildStoredBlockInfoByBlock(height, totalBlocksCommitted, ctx.BlockModel)
		if err != nil {
			return err
		}
	} else {
		storedBlockInfoList, err = buildStoredBlockInfoByRollup(height, totalBlocksCommitted, cli, ctx.L1RollupTxModel)
		if err != nil {
			return err
		}
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
	logx.Infof("send revert block success,tx hash=%s,startHeight=%d ~ endHeight=%d", txHash, startHeight, totalBlocksCommitted)
	err = checkRevertBlock(cli, c, txHash)
	if err != nil {
		return err
	}
	logx.Infof("revert block success,tx hash=%s,startHeight=%d ~ endHeight=%d", txHash, startHeight, totalBlocksCommitted)
	return nil
}

func buildStoredBlockInfoByBlock(height int64, endHeight int64, blockModel block.BlockModel) ([]zkbnb.StorageStoredBlockInfo, error) {
	if height > endHeight {
		return nil, fmt.Errorf("the latest height of TxTypeCommit is %d,it is less than input param height %d,pls check", endHeight, height)
	}

	storedBlockInfoList := make([]zkbnb.StorageStoredBlockInfo, 0)
	for height <= endHeight {
		blockInfo, err := blockModel.GetBlockByHeightWithoutTx(height)
		if err != nil {
			return nil, fmt.Errorf("failed to get block info, err: %v", err)
		}
		if blockInfo.BlockStatus != block.StatusCommitted {
			return nil, fmt.Errorf("invalid block status, blockHeight=%d,status=%d", height, blockInfo.BlockStatus)
		}
		storedBlockInfo := chain.ConstructStoredBlockInfo(blockInfo)
		storedBlockInfoList = append(storedBlockInfoList, storedBlockInfo)
		height++
	}
	return storedBlockInfoList, nil
}

func buildStoredBlockInfoByRollup(height int64, endHeight int64, cli *rpc.ProviderClient, l1RollupTxModel l1rolluptx.L1RollupTxModel) ([]zkbnb.StorageStoredBlockInfo, error) {
	storedBlockInfoList := make([]zkbnb.StorageStoredBlockInfo, 0)
	handledTxs, err := l1RollupTxModel.GetHandledCommitTxList(height)
	if err != nil && err != types.DbErrNotFound {
		return nil, fmt.Errorf("getHandledCommitTxList error %v", err)
	}
	for _, handledTx := range handledTxs {
		storedBlockInfo, err := getStoredBlockInfoFromL1(endHeight, cli, handledTx.L1TxHash)
		if err != nil {
			return nil, fmt.Errorf("failed to getStoredBlockInfoFromL1, err: %v", err)
		}

		storedBlockInfoList = append(storedBlockInfoList, storedBlockInfo...)
	}
	return storedBlockInfoList, nil
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

		latestL1Height, err := cli.GetHeight()
		if err != nil {
			return fmt.Errorf("failed to get l1 block height, err: %v", err)
		}
		if latestL1Height < receipt.BlockNumber.Uint64()+c.ChainConfig.ConfirmBlocksCount {
			continue
		} else {
			if receipt.Status == 0 {
				return fmt.Errorf("failed to revert block, tx_hash=%s,receipt.Status=0", txHash)
			}
			return nil
		}
	}
}

func getStoredBlockInfoFromL1(endHeight int64, cli *rpc.ProviderClient, hash string) (storedBlockInfoList []zkbnb.StorageStoredBlockInfo, err error) {
	transaction, _, err := cli.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		return storedBlockInfoList, err
	}

	storageStoredBlockInfo := desertexit.StorageStoredBlockInfo{}
	newBlocksData := make([]desertexit.ZkBNBCommitBlockInfo, 0)
	callData := desertexit.CommitBlocksCallData{LastCommittedBlockData: &storageStoredBlockInfo, NewBlocksData: newBlocksData}
	newABIDecoder := abicoder.NewABIDecoder(monitor2.ZkBNBContractAbi)
	if err := newABIDecoder.UnpackIntoInterface(&callData, "commitBlocks", transaction.Data()[4:]); err != nil {
		return storedBlockInfoList, err
	}
	stateRootMap := make(map[uint32]string, 0)
	for _, commitBlockInfo := range callData.NewBlocksData {
		stateRootMap[commitBlockInfo.BlockNumber] = common.Bytes2Hex(commitBlockInfo.NewStateRoot[:])
	}
	stateRootMap[callData.LastCommittedBlockData.BlockNumber] = common.Bytes2Hex(callData.LastCommittedBlockData.StateRoot[:])

	for _, commitBlockInfo := range callData.NewBlocksData {
		if int64(commitBlockInfo.BlockNumber) > endHeight {
			continue
		}

		//if !(commitBlockInfo.BlockNumber <= 4576 && commitBlockInfo.BlockNumber >= 4576) {
		//	continue
		//}
		commitment := chain.CreateBlockCommitment(int64(commitBlockInfo.BlockNumber), commitBlockInfo.Timestamp.Int64(),
			common.FromHex(stateRootMap[commitBlockInfo.BlockNumber-1]), common.FromHex(common.Bytes2Hex(commitBlockInfo.NewStateRoot[:])),
			commitBlockInfo.PublicData, int64(len(getPubDataOffset(commitBlockInfo.OnchainOperations))))
		var (
			commitmentByte               [32]byte
			pendingOnChainOperationsHash [32]byte
		)
		copy(commitmentByte[:], common.FromHex(commitment)[:])
		onChainOperationsHash, priorityOperations := getPendingOnChainOperationsHash(commitBlockInfo.PublicData)
		copy(pendingOnChainOperationsHash[:], onChainOperationsHash[:])

		logx.Infof("getStoredBlockInfoFromL1,BlockNumber=%d, commitment=%s,priorityOperations=%d,pendingOnChainOperationsHash=%s", commitBlockInfo.BlockNumber, commitment, priorityOperations, common.Bytes2Hex(pendingOnChainOperationsHash[:]))

		storedBlockInfo := zkbnb.StorageStoredBlockInfo{
			BlockSize:                    commitBlockInfo.BlockSize,
			BlockNumber:                  commitBlockInfo.BlockNumber,
			PriorityOperations:           uint64(priorityOperations),
			PendingOnchainOperationsHash: pendingOnChainOperationsHash,
			Timestamp:                    commitBlockInfo.Timestamp,
			StateRoot:                    commitBlockInfo.NewStateRoot,
			Commitment:                   commitmentByte,
		}
		storedBlockInfoList = append(storedBlockInfoList, storedBlockInfo)
	}
	return storedBlockInfoList, nil
}

func getPubDataOffset(onChainOperations []desertexit.ZkBNBOnchainOperationData) []uint32 {
	publicDataOffsetList := make([]uint32, 0)
	for _, onChainOperation := range onChainOperations {
		publicDataOffsetList = append(publicDataOffsetList, onChainOperation.PublicDataOffset)
	}
	return publicDataOffsetList
}

func getPendingOnChainOperationsHash(pubData []byte) ([]byte, int) {
	priorityOperations := 0
	pendingOnChainOperationsHash := common.FromHex(types.EmptyStringKeccak)
	sizePerTx := types2.PubDataBitsSizePerTx / 8
	for i := 0; i < len(pubData)/sizePerTx; i++ {
		subPubData := pubData[i*sizePerTx : (i+1)*sizePerTx]
		offset := 0
		offset, txType := common2.ReadUint8(subPubData, offset)
		if types.IsPriorityOperationTx(int64(txType)) {
			priorityOperations++
		}
		if int64(txType) == types.TxTypeFullExitNft ||
			int64(txType) == types.TxTypeFullExit ||
			int64(txType) == types.TxTypeWithdrawNft ||
			int64(txType) == types.TxTypeWithdraw {
			pendingOnChainOperationsHash = common2.ConcatKeccakHash(pendingOnChainOperationsHash, subPubData)
		}
	}
	return pendingOnChainOperationsHash, priorityOperations
}

func getTotalBlocksCommitted(cli *rpc.ProviderClient, zkBnbContract string) (int64, error) {
	zkBnbInstance, err := zkbnb.LoadZkBNBInstance(cli, zkBnbContract)
	if err != nil {
		logx.Severef("failed toLoadZkBNBInstance, %v", err)
		return 0, err
	}
	totalBlocksCommitted, err := zkBnbInstance.TotalBlocksCommitted(&bind.CallOpts{})
	if err != nil {
		logx.Severef("failed TotalBlocksCommitted, %v", err)
		return 0, err
	}
	logx.Infof("TotalBlocksCommitted=%d", totalBlocksCommitted)
	return int64(totalBlocksCommitted), nil
}
