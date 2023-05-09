/*
 * Copyright © 2021 ZkBNB Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/common/log"
	"github.com/bnb-chain/zkbnb/core/rpc_client"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
	"gorm.io/plugin/dbresolver"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	sconfig "github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	MaxErrorRetryCount = 3
	// VM Exception while processing transaction: revert i
	CallContractError = "revert"

	RpcOverSized = "oversized data"

	ReplacementTransactionUnderpriced = "replacement transaction underpriced"

	TransactionUnderpriced = "transaction underpriced"
)

var (
	l2BlockCommitToChainHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_commit_to_chain_height",
		Help:      "l2Block_roll_up_height metrics.",
	})

	l2BlockCommitConfirmByChainHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_commit_confirm_by_chain_height",
		Help:      "l2Block_roll_up_height metrics.",
	})

	l2BlockSubmitToVerifyHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_submit_to_verify_height",
		Help:      "l2Block_roll_up_height metrics.",
	})

	l2BlockVerifiedHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_verified_height",
		Help:      "l2Block_roll_up_height metrics.",
	})
	l2MaxWaitingTimeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_max_waiting_time",
		Help:      "l2Block_roll_up_time metrics.",
	})
	l1HeightSenderMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1Block_block_height_send",
		Help:      "l1Block_block_height_send metrics.",
	})
	l1ExceptionSenderMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_Exception_send",
		Help:      "l1_Exception_send metrics.",
	})
	commitExceptionHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_Exception_height",
		Help:      "commit_Exception_height metrics.",
	})
	verifyExceptionHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verify_Exception_height",
		Help:      "verify_Exception_height metrics.",
	})
	contractBalanceMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "contract_balance",
			Help:      "contract_balance metrics.",
		},
		[]string{"type"})
	batchCommitContactMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "batch_commit_contact",
			Help:      "batch_commit_contact metrics.",
		},
		[]string{"type"})
	batchCommitCostMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "batch_commit_cost",
		Help:      "batch_commit_cost metrics.",
	})
	batchVerifyContactMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "batch_verify_contact",
			Help:      "batch_verify_contact metrics.",
		},
		[]string{"type"})
	batchVerifyCostMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "batch_verify_cost",
		Help:      "batch_verify_cost metrics.",
	})
	batchTotalCostMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "batch_total_cost",
		Help:      "batch_total_cost metrics.",
	})
	runTimeIntervalMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "run_time_interval",
			Help:      "run_time_interval metrics.",
		},
		[]string{"type"})
)

type CommitBlockData struct {
	compressedBlocks    []*compressedblock.CompressedBlock
	commitBlockList     []zkbnb.ZkBNBCommitBlockInfo
	lastStoredBlockInfo zkbnb.StorageStoredBlockInfo
	totalEstimatedFee   uint64
	maxGasPrice         decimal.Decimal
	gasPrice            *big.Int
	nonce               uint64
}

type VerifyAndExecuteBlockData struct {
	verifyAndExecuteBlocksInfo []zkbnb.ZkBNBVerifyAndExecuteBlockInfo
	blocks                     []*block.Block
	proofs                     []*big.Int
	totalEstimatedFee          uint64
	maxGasPrice                decimal.Decimal
	gasPrice                   *big.Int
	nonce                      uint64
}

type Sender struct {
	config               sconfig.Config
	goCache              *ristretto.Cache
	ZkBNBContractAddress string

	kmsClient          *kms.Client
	commitAuthClient   *rpc.AuthClient
	verifyAuthClient   *rpc.AuthClient
	commitKmsKeyClient *rpc.KMSKeyClient
	verifyKmsKeyClient *rpc.KMSKeyClient

	zkbnbClient *zkbnb.ZkBNBClient

	// Data access objects
	db                       *gorm.DB
	blockModel               block.BlockModel
	compressedBlockModel     compressedblock.CompressedBlockModel
	l1RollupTxModel          l1rolluptx.L1RollupTxModel
	sysConfigModel           sysconfig.SysConfigModel
	proofModel               proof.ProofModel
	txModel                  tx.TxModel
	accountModel             account.AccountModel
	metricCommitRollupTxId   uint
	metricCommitRollupHeight int64
	metricVerifyRollupTxId   uint
	metricVerifyRollupHeight int64
}

func NewSender(c sconfig.Config) *Sender {

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))
	if c.ChainConfig.MaxGasPriceIncreasePercentage == 0 {
		// Calculation Formula:Percentage = ((MaxGasPrice-GasPrice)/GasPrice)*100
		c.ChainConfig.MaxGasPriceIncreasePercentage = 50
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000000,
		MaxCost:     100000,
		BufferItems: 64, // official recommended value

		// Called when setting cost to 0 in `Set/SetWithTTL`
		Cost: func(value interface{}) int64 {
			return 1
		},
		OnEvict: func(item *ristretto.Item) {
			//logx.Infof("OnEvict %d, %d, %v, %v", item.Key, item.Cost, item.Value, item.Expiration)
		},
	})
	if err != nil {
		logx.Severe("MemCache init failed")
		panic("MemCache init failed")
	}

	s := &Sender{
		config:                   c,
		goCache:                  cache,
		db:                       db,
		blockModel:               block.NewBlockModel(db),
		compressedBlockModel:     compressedblock.NewCompressedBlockModel(db),
		l1RollupTxModel:          l1rolluptx.NewL1RollupTxModel(db),
		sysConfigModel:           sysconfig.NewSysConfigModel(db),
		proofModel:               proof.NewProofModel(db),
		txModel:                  tx.NewTxModel(db),
		accountModel:             account.NewAccountModel(db),
		metricCommitRollupTxId:   0,
		metricCommitRollupHeight: 0,
		metricVerifyRollupTxId:   0,
		metricVerifyRollupHeight: 0,
	}

	rollupAddress, err := s.sysConfigModel.GetSysConfigByName(types.ZkBNBContract)
	if err != nil {
		logx.Severef("fatal error, failed to get zkBNB contract configuration, err:%v, SysConfigName:%s",
			err, types.ZkBNBContract)
		panic("fatal error, failed to get zkBNB contract configuration, err:" + err.Error() + "SysConfigName:" +
			types.ZkBNBContract)
	}
	s.ZkBNBContractAddress = rollupAddress.Value

	err = rpc_client.InitRpcClients(s.sysConfigModel, c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("failed to create rpc client instance, %v", err)
		panic("failed to create rpc client instance, err:" + err.Error())
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logx.Severef("failed to load KMS client config, %v", err)
		panic("failed to load KMS client config, err:" + err.Error())
	}
	s.kmsClient = kms.NewFromConfig(cfg)

	chainId, err := s.getProviderClient().ChainID(context.Background())
	if err != nil {
		logx.Severef("fatal error, failed to get the chainId from the l1 server, err:%v", err)
		panic("fatal error, failed to get the chainId from the l1 server, err:" + err.Error())
	}

	commitKeyId := c.KMSConfig.CommitKeyId
	verifyKeyId := c.KMSConfig.VerifyKeyId
	if len(commitKeyId) > 0 && len(verifyKeyId) > 0 {
		s.commitKmsKeyClient, err = rpc.NewKMSKeyClient(s.kmsClient, c.KMSConfig.CommitKeyId, chainId)
		if err != nil {
			logx.Severef("fatal error, failed to initiate commit kmsKeyClient instance, err:%v", err)
			panic("fatal error, failed to initiate commit kmsKeyClient instance, err:" + err.Error())
		}

		s.verifyKmsKeyClient, err = rpc.NewKMSKeyClient(s.kmsClient, c.KMSConfig.VerifyKeyId, chainId)
		if err != nil {
			logx.Severef("fatal error, failed to initiate verify kmsKeyClient instance, err:%v", err)
			panic("fatal error, failed to initiate verify kmsKeyClient instance, err:" + err.Error())
		}
	} else {
		commitBlockSk := c.AuthConfig.CommitBlockSk
		verifyBlockSk := c.AuthConfig.VerifyBlockSk
		if len(commitBlockSk) > 0 && len(verifyBlockSk) > 0 {
			s.commitAuthClient, err = rpc.NewAuthClient(commitBlockSk, chainId)
			if err != nil {
				logx.Severef("fatal error, failed to initiate commit authClient instance, err:%v", err)
				panic("fatal error, failed to initiate commit authClient instance, err:" + err.Error())
			}

			s.verifyAuthClient, err = rpc.NewAuthClient(verifyBlockSk, chainId)
			if err != nil {
				logx.Severef("fatal error, failed to initiate verify authClient instance, err:%v", err)
				panic("fatal error, failed to initiate verify authClient instance, err:" + err.Error())
			}
		} else {
			logx.Severef("fatal error, both kms keys and auth private keys not set!")
			panic("fatal error, both kms keys and auth private keys not set!")
		}
	}

	commitConstructor, err := s.GenerateConstructorForCommit()
	if err != nil {
		logx.Severef("fatal error, GenerateConstructorForCommit raises error:%v", err)
		panic("fatal error, GenerateConstructorForCommit raises error:" + err.Error())
	}
	verifyConstructor, err := s.GenerateConstructorForVerifyAndExecute()
	if err != nil {
		logx.Severef("fatal error, GenerateConstructorForVerifyAndExecute raises error:%v", err)
		panic("fatal error, GenerateConstructorForVerifyAndExecute raises error:" + err.Error())
	}

	s.zkbnbClient, err = zkbnb.NewZkBNBClient(s.getProviderClient(), rollupAddress.Value)
	s.zkbnbClient.CommitConstructor = commitConstructor
	s.zkbnbClient.VerifyConstructor = verifyConstructor
	if err != nil {
		logx.Severef("fatal error, ZkBNBClient initiate raises error:%v", err)
		panic("fatal error, ZkBNBClient initiate raises error:" + err.Error())
	}

	return s
}

func InitPrometheusFacility() {
	if err := prometheus.Register(l2BlockCommitToChainHeightMetric); err != nil {
		logx.Errorf("prometheus.Register l2BlockCommitToChainHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockCommitConfirmByChainHeightMetric); err != nil {
		logx.Errorf("prometheus.Register l2BlockCommitConfirmByChainHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockSubmitToVerifyHeightMetric); err != nil {
		logx.Errorf("prometheus.Register l2BlockSubmitToVerifyHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockVerifiedHeightMetric); err != nil {
		logx.Errorf("prometheus.Register l2BlockVerifiedHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2MaxWaitingTimeMetric); err != nil {
		logx.Errorf("prometheus.Register l2MaxWaitingTimeMetric error: %v", err)
	}
	if err := prometheus.Register(l1HeightSenderMetric); err != nil {
		logx.Errorf("prometheus.Register l1HeightSenderMetric error: %v", err)
	}
	if err := prometheus.Register(l1ExceptionSenderMetric); err != nil {
		logx.Errorf("prometheus.Register l1ExceptionSenderMetric error: %v", err)
	}
	if err := prometheus.Register(commitExceptionHeightMetric); err != nil {
		logx.Errorf("prometheus.Register commitExceptionHeightMetric error: %v", err)
	}
	if err := prometheus.Register(verifyExceptionHeightMetric); err != nil {
		logx.Errorf("prometheus.Register verifyExceptionHeightMetric error: %v", err)
	}
	if err := prometheus.Register(contractBalanceMetric); err != nil {
		logx.Errorf("prometheus.Register contractBalanceMetric error: %v", err)
	}
	if err := prometheus.Register(batchCommitContactMetric); err != nil {
		logx.Errorf("prometheus.Register batchCommitContactMetric error: %v", err)
	}
	if err := prometheus.Register(batchCommitCostMetric); err != nil {
		logx.Errorf("prometheus.Register batchCommitCostMetric error: %v", err)
	}
	if err := prometheus.Register(batchVerifyContactMetric); err != nil {
		logx.Errorf("prometheus.Register batchVerifyContactMetric error: %v", err)
	}
	if err := prometheus.Register(batchVerifyCostMetric); err != nil {
		logx.Errorf("prometheus.Register batchVerifyCostMetric error: %v", err)
	}
	if err := prometheus.Register(batchTotalCostMetric); err != nil {
		logx.Errorf("prometheus.Register batchTotalCostMetric error: %v", err)
	}
	if err := prometheus.Register(runTimeIntervalMetric); err != nil {
		logx.Errorf("prometheus.Register runTimeIntervalMetric error: %v", err)
	}

}

func (s *Sender) CommitBlocks() (err error) {

	exist, err := s.ExistPendingTx(l1rolluptx.TxTypeCommit)
	if err != nil {
		return err
	}
	// If there exists pending transaction, do not commit and directly return
	if exist {
		return nil
	}

	lastHandledTx, err := s.PrepareLastHandledTx(l1rolluptx.TxTypeCommit)
	if err != nil {
		return err
	}

	commitBlockData, shouldCommit, err := s.PrepareCommitBlockData(lastHandledTx)
	if err != nil {
		return err
	}
	if commitBlockData == nil {
		return nil
	}

	lastStoredBlockInfo := commitBlockData.lastStoredBlockInfo
	pendingCommitBlocks := commitBlockData.commitBlockList
	compressedBlocks := commitBlockData.compressedBlocks
	maxGasPrice := commitBlockData.maxGasPrice
	gasPrice := commitBlockData.gasPrice
	nonce := commitBlockData.nonce

	if len(compressedBlocks) == 0 {
		return
	}

	l2BlockHeight := int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber)
	ctx := log.NewCtxWithKV(log.BlockHeightContext, l2BlockHeight)

	// Judge whether the blocks should be committed to the chain for better gas consumption
	if !shouldCommit {
		logx.Infof("check again if need to ShouldCommitBlocks,l2BlockHeight=%d", l2BlockHeight)
		shouldCommit = s.ShouldCommitBlocks(compressedBlocks, ctx)
		if !shouldCommit {
			logx.WithContext(ctx).Infof("abandon commit block to l1")
			return nil
		}
	}

	var txHash string
	retry := false
	for {
		if retry {
			newNonce, err := s.getProviderClient().GetPendingNonce(s.GetCommitAddress().Hex())
			if err != nil {
				return fmt.Errorf("failed to get nonce for commit block, errL %v", err)
			}
			if nonce != newNonce {
				return fmt.Errorf("failed to retry for commit block, nonce=%d,newNonce=%d", nonce, newNonce)
			}
			standByGasPrice := decimal.NewFromInt(gasPrice.Int64()).Add(decimal.NewFromInt(gasPrice.Int64()).Mul(decimal.NewFromFloat(0.1)))
			if standByGasPrice.GreaterThan(maxGasPrice) {
				logx.WithContext(ctx).Errorf("abandon commit block to l1, gasPrice>maxGasPrice,l1 nonce: %d,gasPrice: %d,maxGasPrice: %d", nonce, standByGasPrice, maxGasPrice)
				return nil
			}
			gasPrice = standByGasPrice.RoundUp(0).BigInt()
			logx.WithContext(ctx).Infof("speed up commit block to l1,l1 nonce: %d,gasPrice: %d", nonce, gasPrice)
		}

		s.ValidOverSuggestGasPrice150Percent(gasPrice)

		// commit blocks on-chain
		l2BlockHeight = int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber)
		if s.IsOverMaxErrorRetryCount(int64(l2BlockHeight), l1rolluptx.TxTypeCommit) {
			return fmt.Errorf("Send tx to L1 has been called %d times, no more retries,please check.L2BlockHeight=%d,txType=%d ", MaxErrorRetryCount, l2BlockHeight, l1rolluptx.TxTypeCommit)
		}
		txHash, err = s.zkbnbClient.CommitBlocksWithNonce(
			lastStoredBlockInfo,
			pendingCommitBlocks,
			gasPrice,
			s.config.ChainConfig.GasLimit, nonce)
		if err != nil {
			commitExceptionHeightMetric.Set(float64(l2BlockHeight))
			if err.Error() == ReplacementTransactionUnderpriced || err.Error() == TransactionUnderpriced {
				logx.WithContext(ctx).Errorf("failed to send commit tx,try again: errL %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
				retry = true
				continue
			}
			if err.Error() == RpcOverSized {
				logx.WithContext(ctx).Errorf("failed to send commit tx,try again after deleting one block: errL %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
				pendingCommitBlocks = pendingCommitBlocks[0 : len(pendingCommitBlocks)-1]
				continue
			}
			s.HandleSendTxToL1Error(int64(l2BlockHeight), l1rolluptx.TxTypeCommit, txHash, err)
			return fmt.Errorf("failed to send commit tx, errL %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
		}
		break
	}

	commitExceptionHeightMetric.Set(float64(0))
	for _, pendingCommitBlock := range pendingCommitBlocks {
		l2BlockCommitToChainHeightMetric.Set(float64(pendingCommitBlock.BlockNumber))
	}
	newRollupTx := &l1rolluptx.L1RollupTx{
		L1TxHash:      txHash,
		TxStatus:      l1rolluptx.StatusPending,
		TxType:        l1rolluptx.TxTypeCommit,
		L2BlockHeight: l2BlockHeight,
		L1Nonce:       int64(nonce),
		GasPrice:      gasPrice.Int64(),
	}
	err = s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
	if err != nil {
		return fmt.Errorf("failed to create tx in database, err: %v", err)
	}
	l2BlockCommitToChainHeightMetric.Set(float64(newRollupTx.L2BlockHeight))
	logx.Infof("new blocks have been committed(height): %v:%s", newRollupTx.L2BlockHeight, newRollupTx.L1TxHash)
	return nil
}

func (s *Sender) UpdateSentTxs() (err error) {
	pendingTxs, err := s.l1RollupTxModel.GetL1RollupTxsByStatus(l1rolluptx.StatusPending)
	if err != nil {
		if err == types.DbErrNotFound {
			return nil
		}
		return fmt.Errorf("failed to get pending txs, err: %v", err)
	}
	cli := s.getProviderClient()
	latestL1Height, err := cli.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get l1 block height, err: %v", err)
	}
	l1HeightSenderMetric.Set(float64(latestL1Height))
	var (
		pendingUpdateRxs         []*l1rolluptx.L1RollupTx
		pendingUpdateProofStatus = make(map[int64]int)
	)
	for _, pendingTx := range pendingTxs {
		txHash := pendingTx.L1TxHash
		receipt, err := cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("query transaction receipt %s failed, err: %v", txHash, err)
			if time.Now().After(pendingTx.UpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				// No need to check the response, do best effort.
				logx.Errorf("delete timeout l1 rollup tx, tx_hash=%s", pendingTx.L1TxHash)
				//nolint:errcheck
				s.l1RollupTxModel.DeleteL1RollupTx(pendingTx)
				l2MaxWaitingTimeMetric.Set(float64(pendingTx.L2BlockHeight))
			}
			continue
		}
		if receipt.Status == 0 {
			// Should direct mark tx deleted
			l1ExceptionSenderMetric.Set(float64(pendingTx.L2BlockHeight))
			logx.Severef("Transaction failed to execute on L1: %v", txHash)
			cacheKey := fmt.Sprintf("%s-%d-%d", SentBlockToL1ErrorPrefix, pendingTx.TxType, pendingTx.L2BlockHeight)
			retryCount := 0
			cacheValue, found := s.goCache.Get(cacheKey)
			if found {
				retryCount = cacheValue.(int)
				if retryCount >= MaxErrorRetryCount {
					logx.Severef("Commit to L1 has been retried %d times, no more retries,txHash=%s,L2BlockHeight=%d", retryCount, txHash, pendingTx.L2BlockHeight)
					continue
				}
				s.goCache.SetWithTTL(cacheKey, retryCount+1, 0, time.Minute*120)
			} else {
				retryCount = 1
				s.goCache.SetWithTTL(cacheKey, retryCount, 0, time.Minute*120)
			}
			logx.Infof("Commit to L1 has been retried %d times,txHash=%s,L2BlockHeight=%d", retryCount, txHash, pendingTx.L2BlockHeight)
			logx.Infof("delete timeout l1 rollup tx, tx_hash=%s", pendingTx.L1TxHash)
			s.l1RollupTxModel.DeleteL1RollupTx(pendingTx)
			continue
		}
		l2MaxWaitingTimeMetric.Set(float64(0))
		l1ExceptionSenderMetric.Set(float64(0))

		// not finalized yet
		if latestL1Height < receipt.BlockNumber.Uint64()+s.config.ChainConfig.ConfirmBlocksCount {
			continue
		}
		var validTx bool
		for _, vlog := range receipt.Logs {
			switch vlog.Topics[0].Hex() {
			case zkbnbLogBlockCommitSigHash.Hex():
				var event zkbnb.ZkBNBBlockCommit
				if err = ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
					return err
				}
				validTx = int64(event.BlockNumber) == pendingTx.L2BlockHeight
			case zkbnbLogBlockVerificationSigHash.Hex():
				var event zkbnb.ZkBNBBlockVerification
				if err = ZkBNBContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
					return err
				}
				validTx = int64(event.BlockNumber) == pendingTx.L2BlockHeight
				pendingUpdateProofStatus[int64(event.BlockNumber)] = proof.Confirmed
			case zkbnbLogBlocksRevertSigHash.Hex():
				// TODO revert
			default:
			}
		}

		if validTx {
			pendingTx.TxStatus = l1rolluptx.StatusHandled
			pendingTx.GasUsed = receipt.GasUsed
			pendingUpdateRxs = append(pendingUpdateRxs, pendingTx)
			if pendingTx.TxType == l1rolluptx.TxTypeCommit {
				l2BlockCommitConfirmByChainHeightMetric.Set(float64(pendingTx.L2BlockHeight))
			} else if pendingTx.TxType == l1rolluptx.TxTypeVerifyAndExecute {
				l2BlockVerifiedHeightMetric.Set(float64(pendingTx.L2BlockHeight))
			}
		}
	}

	//update db
	err = s.db.Transaction(func(tx *gorm.DB) error {
		//update l1 rollup txs
		err := s.l1RollupTxModel.UpdateL1RollupTxsStatusInTransact(tx, pendingUpdateRxs)
		if err != nil {
			return err
		}
		//update proof status
		err = s.proofModel.UpdateProofsInTransact(tx, pendingUpdateProofStatus)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to updte rollup txs, err:%v", err)
	}
	return nil
}

func (s *Sender) VerifyAndExecuteBlocks() (err error) {

	exist, err := s.ExistPendingTx(l1rolluptx.TxTypeVerifyAndExecute)
	if err != nil {
		return err
	}
	// If there exists pending transaction, do not verify and directly return
	if exist {
		return nil
	}

	lastHandledTx, err := s.PrepareLastHandledTx(l1rolluptx.TxTypeVerifyAndExecute)
	if err != nil {
		return err
	}

	verifyAndExecuteBlockData, shouldVerifyAndExecute, err := s.PrepareVerifyAndExecuteBlockData(lastHandledTx)
	if err != nil {
		return err
	}
	if verifyAndExecuteBlockData == nil {
		return nil
	}

	nonce := verifyAndExecuteBlockData.nonce
	blocks := verifyAndExecuteBlockData.blocks
	proofs := verifyAndExecuteBlockData.proofs
	gasPrice := verifyAndExecuteBlockData.gasPrice
	maxGasPrice := verifyAndExecuteBlockData.maxGasPrice
	pendingVerifyAndExecuteBlocks := verifyAndExecuteBlockData.verifyAndExecuteBlocksInfo

	if len(blocks) == 0 {
		return nil
	}

	l2BlockHeight := blocks[len(blocks)-1].BlockHeight
	ctx := log.NewCtxWithKV(log.BlockHeightContext, l2BlockHeight)

	// Judge whether the blocks should be verified and executed to the chain for better gas consumption
	if !shouldVerifyAndExecute {
		logx.Infof("check again if need to ShouldVerifyAndExecuteBlocks,l2BlockHeight=%d", l2BlockHeight)
		shouldVerifyAndExecute := s.ShouldVerifyAndExecuteBlocks(blocks, ctx)
		if !shouldVerifyAndExecute {
			logx.WithContext(ctx).Infof("abandon verify and execute block to l1")
			return nil
		}
	}

	txHash := ""
	retry := false
	for {
		if retry {
			newNonce, err := s.getProviderClient().GetPendingNonce(s.GetVerifyAddress().Hex())
			if err != nil {
				return fmt.Errorf("failed to get nonce for verify block, errL %v", err)
			}
			if nonce != newNonce {
				return fmt.Errorf("failed to retry for verify block, nonce=%d,newNonce=%d", nonce, newNonce)
			}
			standByGasPrice := decimal.NewFromInt(gasPrice.Int64()).Add(decimal.NewFromInt(gasPrice.Int64()).Mul(decimal.NewFromFloat(0.1)))
			if standByGasPrice.GreaterThan(maxGasPrice) {
				logx.WithContext(ctx).Errorf("abandon verify block to l1, gasPrice>maxGasPrice,l1 nonce: %d,gasPrice: %d,maxGasPrice: %d", nonce, standByGasPrice, maxGasPrice)
				return nil
			}
			gasPrice = standByGasPrice.RoundUp(0).BigInt()
			logx.WithContext(ctx).Infof("speed up verify block to l1,l1 nonce: %d,gasPrice: %d", nonce, gasPrice)
		}

		s.ValidOverSuggestGasPrice150Percent(gasPrice)

		// Verify blocks on-chain
		l2BlockHeight = int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber)
		if s.IsOverMaxErrorRetryCount(l2BlockHeight, l1rolluptx.TxTypeVerifyAndExecute) {
			return fmt.Errorf("Send tx to L1 has been called %d times, no more retries,please check,L2BlockHeight=%d,txType=TxTypeVerifyAndExecute ", MaxErrorRetryCount, l2BlockHeight)
		}
		txHash, err = s.zkbnbClient.VerifyAndExecuteBlocksWithNonce(
			pendingVerifyAndExecuteBlocks,
			proofs, gasPrice, s.config.ChainConfig.GasLimit, nonce)
		if err != nil {
			verifyExceptionHeightMetric.Set(float64(l2BlockHeight))
			if err.Error() == ReplacementTransactionUnderpriced || err.Error() == TransactionUnderpriced {
				logx.WithContext(ctx).Errorf("failed to send verify tx,try again: errL %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
				retry = true
				continue
			}
			if err.Error() == RpcOverSized {
				logx.WithContext(ctx).Errorf("failed to send verify tx,try again after deleting one block: errL %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
				pendingVerifyAndExecuteBlocks = pendingVerifyAndExecuteBlocks[0 : len(pendingVerifyAndExecuteBlocks)-1]
				continue
			}
			s.HandleSendTxToL1Error(l2BlockHeight, l1rolluptx.TxTypeVerifyAndExecute, txHash, err)
			return fmt.Errorf("failed to send verify tx: %v:%s,blockHeight=%d,nonce=%d,gasPrice=%s", err, txHash, l2BlockHeight, nonce, gasPrice.String())
		}
		break
	}

	verifyExceptionHeightMetric.Set(float64(0))
	for _, pendingVerifyAndExecuteBlock := range pendingVerifyAndExecuteBlocks {
		l2BlockSubmitToVerifyHeightMetric.Set(float64(pendingVerifyAndExecuteBlock.BlockHeader.BlockNumber))
	}
	newRollupTx := &l1rolluptx.L1RollupTx{
		L1TxHash:      txHash,
		TxStatus:      l1rolluptx.StatusPending,
		TxType:        l1rolluptx.TxTypeVerifyAndExecute,
		L2BlockHeight: l2BlockHeight,
		L1Nonce:       int64(nonce),
		GasPrice:      gasPrice.Int64(),
	}
	err = s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to create rollup tx in db %v", err))
	}
	l2BlockSubmitToVerifyHeightMetric.Set(float64(newRollupTx.L2BlockHeight))
	logx.Infof("new blocks have been verified and executed(height): %d:%s", newRollupTx.L2BlockHeight, newRollupTx.L1TxHash)
	return nil
}

func (s *Sender) PrepareCommitBlockData(lastHandledTx *l1rolluptx.L1RollupTx) (*CommitBlockData, bool, error) {
	commitTxCountLimit := sconfig.GetSenderConfig().CommitTxCountLimit
	maxCommitBlockCount := sconfig.GetSenderConfig().MaxCommitBlockCount
	maxCommitTotalGasFee := sconfig.GetSenderConfig().MaxCommitTotalGasFee

	var start = int64(1)
	var totalTxCount uint64 = 0
	var commitBlockData = &CommitBlockData{}

	if lastHandledTx != nil {
		latestCommittedHeight, err := s.blockModel.GetLatestCommittedHeight()
		if err != nil && err != types.DbErrNotFound {
			return nil, false, err
		}
		latestVerifiedHeight, err := s.blockModel.GetLatestVerifiedHeight()
		if err != nil && err != types.DbErrNotFound {
			return nil, false, err
		}
		lastHandledTx.L2BlockHeight = common2.MaxInt64(common2.MaxInt64(latestCommittedHeight, lastHandledTx.L2BlockHeight), latestVerifiedHeight)
		start = lastHandledTx.L2BlockHeight + 1
	}

	for {
		compressedBlocks, err := s.compressedBlockModel.GetCompressedBlocksBetween(start, start+int64(maxCommitBlockCount))
		if err != nil && err != types.DbErrNotFound {
			return nil, false, fmt.Errorf("failed to get compress block err: %v", err)
		}
		logx.Infof("GetCompressedBlocksForCommit: start:%d, maxCommitBlockCount:%d, compressed block count:%d",
			start, maxCommitBlockCount, len(compressedBlocks))
		if len(compressedBlocks) == 0 {
			return nil, false, nil
		}

		commitBlockList, err := ConvertBlocksForCommitToCommitBlockInfos(compressedBlocks, s.txModel)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get commit block info, err: %v", err)
		}

		lastStoredBlockInfo, err := s.PrepareLastStoredBlockInfo(lastHandledTx)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get last stored block info, err: %v", err)
		}

		l2BlockHeight := int64(commitBlockList[len(commitBlockList)-1].BlockNumber)
		ctx := log.NewCtxWithKV(log.BlockHeightContext, l2BlockHeight)

		gasPrice, maxGasPrice, err := s.PrepareCommitGasPriceData(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get commit gas price data, err: %v", err)
		}

		nonce, err := s.PrepareCommitNonceValue()
		if err != nil {
			return nil, false, fmt.Errorf("failed to get commit nonce value, err: %v", err)
		}

		// Judge the average tx gas consumption for the committing operation, if the average tx gas consumption is greater
		// than the maxCommitAvgUnitGas, abandon commit operation for temporary
		totalEstimatedFee, err := s.zkbnbClient.EstimateCommitGasWithNonce(lastStoredBlockInfo, commitBlockList, gasPrice, 0, nonce)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get estimated gas fee for committing,last stored block L2BlockHeight=%d err: %v", lastStoredBlockInfo.BlockNumber, err)
		}

		totalTxCount = s.CalculateTotalTxCountForCompressBlock(compressedBlocks)

		commitBlockData.compressedBlocks = compressedBlocks
		commitBlockData.commitBlockList = commitBlockList
		commitBlockData.lastStoredBlockInfo = lastStoredBlockInfo
		commitBlockData.totalEstimatedFee = totalEstimatedFee
		commitBlockData.maxGasPrice = maxGasPrice
		commitBlockData.gasPrice = gasPrice
		commitBlockData.nonce = nonce

		if !sconfig.GetSenderConfig().CommitControlSwitch {
			return commitBlockData, true, nil
		}

		if totalTxCount <= commitTxCountLimit && totalEstimatedFee <= maxCommitTotalGasFee {
			shouldCommit := maxCommitBlockCount != sconfig.GetSenderConfig().MaxCommitBlockCount
			logx.WithContext(ctx).Infof("PrepareCommitBlockData start height=%d,end height=%d,shouldCommit=%s,totalTxCount=%d,commitTxCountLimit=%d,totalEstimatedFee=%d,maxCommitTotalGasFee=%d", start, start+int64(maxCommitBlockCount), shouldCommit, totalTxCount, commitTxCountLimit, totalEstimatedFee, maxCommitTotalGasFee)
			return commitBlockData, shouldCommit, nil
		}

		if maxCommitBlockCount > 1 {
			maxCommitBlockCount--
		} else {
			return nil, false, nil
		}
	}
}

func (s *Sender) ExistPendingTx(txType int64) (bool, error) {
	pendingTx, err := s.l1RollupTxModel.GetLatestPendingTx(txType)
	if err != nil && err != types.DbErrNotFound {
		return false, err
	}
	if pendingTx != nil {
		return true, nil
	}
	return false, nil
}

func (s *Sender) PrepareLastHandledTx(txType int64) (*l1rolluptx.L1RollupTx, error) {
	lastHandledTx, err := s.l1RollupTxModel.GetLatestHandledTx(txType)
	if err != nil && err != types.DbErrNotFound {
		return nil, err
	}
	return lastHandledTx, nil
}

func (s *Sender) PrepareLastStoredBlockInfo(lastHandledTx *l1rolluptx.L1RollupTx) (zkbnb.StorageStoredBlockInfo, error) {
	// get last block info
	lastStoredBlockInfo := DefaultBlockHeader()
	if lastHandledTx != nil {
		lastHandledBlockInfo, err := s.blockModel.GetBlockByHeight(lastHandledTx.L2BlockHeight)
		if err != nil {
			return lastStoredBlockInfo, fmt.Errorf("failed to get last stored block info, err: %v", err)
		}
		// construct last stored block header
		lastStoredBlockInfo = chain.ConstructStoredBlockInfo(lastHandledBlockInfo)
	}
	return lastStoredBlockInfo, nil
}

func (s *Sender) PrepareCommitGasPriceData(ctx context.Context) (*big.Int, decimal.Decimal, error) {
	var err error
	var gasPrice *big.Int
	var emptyMaxGasPrice = decimal.NewFromInt(0)

	if s.config.ChainConfig.GasPrice > 0 {
		gasPrice = big.NewInt(int64(s.config.ChainConfig.GasPrice))
		s.ValidOverSuggestGasPrice(gasPrice)
	} else {
		gasPrice, err = s.getProviderClient().SuggestGasPrice(context.Background())
		if err != nil {
			logx.WithContext(ctx).Errorf("failed to fetch gas price: %v", err)
			return nil, emptyMaxGasPrice, err
		}
	}

	maxGasPrice := (decimal.NewFromInt(gasPrice.Int64()).Mul(decimal.
		NewFromInt(int64(s.config.ChainConfig.MaxGasPriceIncreasePercentage))).
		Div(decimal.NewFromInt(100))).Add(decimal.NewFromInt(gasPrice.Int64()))

	nonce, err := s.getProviderClient().GetPendingNonce(s.GetCommitAddress().Hex())
	if err != nil {
		return nil, emptyMaxGasPrice, fmt.Errorf("failed to get nonce for commit block, errL %v", err)
	}

	l1RollupTx, err := s.l1RollupTxModel.GetLatestByNonce(int64(nonce), l1rolluptx.TxTypeCommit)
	if err != nil && err != types.DbErrNotFound {
		return nil, emptyMaxGasPrice, fmt.Errorf("failed to get latest l1 rollup tx by nonce %d, err: %v", nonce, err)
	}
	if l1RollupTx != nil && l1RollupTx.L1Nonce == int64(nonce) {
		standByGasPrice := decimal.NewFromInt(l1RollupTx.GasPrice).Add(decimal.NewFromInt(l1RollupTx.GasPrice).Mul(decimal.NewFromFloat(0.1)))
		if standByGasPrice.GreaterThan(maxGasPrice) {
			logx.WithContext(ctx).Errorf("abandon commit block to l1, gasPrice>maxGasPrice,l1 nonce: %d,gasPrice: %d,maxGasPrice: %d", nonce, standByGasPrice, maxGasPrice)
			return nil, emptyMaxGasPrice, nil
		}
		gasPrice = standByGasPrice.RoundUp(0).BigInt()
		logx.WithContext(ctx).Infof("speed up commit block to l1,l1 nonce: %d,gasPrice: %d", nonce, gasPrice)
	}
	return gasPrice, maxGasPrice, nil
}

func (s *Sender) PrepareCommitNonceValue() (uint64, error) {
	nonce, err := s.getProviderClient().GetPendingNonce(s.GetCommitAddress().Hex())
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce for commit block, errL %v", err)
	}
	return nonce, nil
}

func (s *Sender) ShouldCommitBlocks(blocks []*compressedblock.CompressedBlock, ctx context.Context) bool {
	// If CommitControlSwitch has been switched off, directly does not perform any control
	if !sconfig.GetSenderConfig().CommitControlSwitch {
		return true
	}

	// Judge the count of the blocks waiting to be committed, if the time count is greater
	// than the MaxCommitBlockCount, commit the blocks directly
	maxCommitBlockCount := sconfig.GetSenderConfig().MaxCommitBlockCount
	if uint64(len(blocks)) >= maxCommitBlockCount {
		logx.WithContext(ctx).Infof("Should commit blocks to l1 network, because blocks count >= maxCommitBlockCount,"+
			"blocks count:%d, maxCommitBlockInterval:%d", len(blocks), maxCommitBlockCount)
		return true
	}

	// Judge the tx count waiting to be committed, if the tx count is greater
	// than the maxCommitTxCount, commit the blocks directly
	maxCommitTxCount := sconfig.GetSenderConfig().MaxCommitTxCount
	totalTxCount := s.CalculateTotalTxCountForCompressBlock(blocks)
	if totalTxCount >= maxCommitTxCount {
		logx.WithContext(ctx).Infof("Should commit blocks to l1 network, because totalTxCount >= maxCommitTxCount,"+
			"totalTxCount:%d, maxCommitTxCount:%d", totalTxCount, maxCommitTxCount)
		return true
	}

	// Judge the time interval of the block waiting to be committed, if the time interval is greater
	// than the maxCommitBlockInterval, commit the blocks directly
	maxCommitBlockInterval := sconfig.GetSenderConfig().MaxCommitBlockInterval
	commitBlockInterval := s.CalculateBlockIntervalForCompressedBlock(blocks)
	if commitBlockInterval >= int64(maxCommitBlockInterval) {
		logx.WithContext(ctx).Infof("Should commit blocks to l1 network, because commitBlockInterval >= maxCommitBlockInterval,"+
			"commitBlockInterval:%d, maxCommitBlockInterval:%d", commitBlockInterval, maxCommitBlockInterval)
		return true
	}

	return false
}

func (s *Sender) ShouldVerifyAndExecuteBlocks(blocks []*block.Block, ctx context.Context) bool {
	// If VerifyControlSwitch has been switched off, directly does not perform any control
	if !sconfig.GetSenderConfig().VerifyControlSwitch {
		return true
	}

	// Judge the count of the blocks waiting to be verified and executed, if the time count is greater
	// than the MaxVerifyBlockCount, verify and execute the blocks directly
	maxVerifyBlockCount := sconfig.GetSenderConfig().MaxVerifyBlockCount
	if uint64(len(blocks)) >= maxVerifyBlockCount {
		logx.WithContext(ctx).Infof("Should verify blocks to l1 network, because blocks count >= maxVerifyBlockCount,"+
			"blocks count:%d, maxVerifyBlockCount:%d", len(blocks), maxVerifyBlockCount)
		return true
	}

	// Judge the tx count waiting to be verified and executed, if the tx count is greater
	// than the maxVerifyTxCount, verify and execute the blocks directly
	maxVerifyTxCount := sconfig.GetSenderConfig().MaxVerifyTxCount
	totalTxCount := s.CalculateTotalTxCountForBlock(blocks)
	if totalTxCount >= maxVerifyTxCount {

		logx.WithContext(ctx).Infof("Should verify blocks to l1 network, because totalTxCount >= maxVerifyTxCount,"+
			"totalTxCount:%d, maxVerifyTxCount:%d", totalTxCount, maxVerifyTxCount)

		return true
	}

	// Judge the time interval of the block waiting to be verified and executed, if the time interval is greater
	// than the maxVerifyBlockInterval, verify and execute the blocks directly
	maxVerifyBlockInterval := sconfig.GetSenderConfig().MaxVerifyBlockInterval
	verifyBlockInterval := s.CalculateBlockIntervalForBlock(blocks)
	if verifyBlockInterval >= int64(maxVerifyBlockInterval) {
		logx.WithContext(ctx).Infof("Should verify blocks to l1 network, because verifyBlockInterval >= maxVerifyBlockInterval,"+
			"verifyBlockInterval:%d, maxVerifyBlockInterval:%d", verifyBlockInterval, maxVerifyBlockInterval)
		return true
	}

	return false
}

func (s *Sender) PrepareVerifyAndExecuteBlockData(lastHandledTx *l1rolluptx.L1RollupTx) (*VerifyAndExecuteBlockData, bool, error) {

	verifyTxCountLimit := sconfig.GetSenderConfig().VerifyTxCountLimit
	maxVerifyBlockCount := sconfig.GetSenderConfig().MaxVerifyBlockCount
	maxVerifyTotalGasFee := sconfig.GetSenderConfig().MaxVerifyTotalGasFee

	var start = int64(1)
	var totalTxCount uint64 = 0
	var verifyAndExecuteBlockData = &VerifyAndExecuteBlockData{}

	if lastHandledTx != nil {
		start = lastHandledTx.L2BlockHeight + 1
	}

	for {
		blocks, err := s.blockModel.GetCommittedBlocksBetween(start, start+int64(maxVerifyBlockCount))
		if err != nil && err != types.DbErrNotFound {
			return nil, false, fmt.Errorf("unable to get blocks to prove, err: %v", err)
		}
		logx.Infof("GetBlocksForVerifyAndExecute: start:%d, maxVerifyBlockCount:%d, verify block count:%d", start, maxVerifyBlockCount, len(blocks))
		if len(blocks) == 0 {
			return nil, false, nil
		}

		pendingVerifyAndExecuteBlocks, err := ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
		if err != nil {
			return nil, false, fmt.Errorf("unable to convert blocks to commit block infos: %v", err)
		}
		l2BlockHeight := int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber)
		ctx := log.NewCtxWithKV(log.BlockHeightContext, l2BlockHeight)

		blockProofs, err := s.proofModel.GetProofsBetween(start, start+int64(len(blocks))-1)
		if err != nil {
			if err == types.DbErrNotFound {
				return nil, false, nil
			}
			return nil, false, fmt.Errorf("unable to get proofs, err: %v", err)
		}

		if err = s.CheckBlockAndProofData(blocks, blockProofs); err != nil {
			return nil, false, err
		}

		proofs, err := s.PrepareProofData(blockProofs)
		if err != nil {
			return nil, false, err
		}

		nonce, err := s.PrepareVerifyNonceValue()
		if err != nil {
			return nil, false, err
		}

		gasPrice, maxGasPrice, err := s.PrepareVerifyGasPriceData(ctx)
		if err != nil {
			return nil, false, err
		}

		// Judge the average tx gas consumption for the verifying and executing operation, if the average tx gas consumption is greater
		// than the maxVerifyAvgUnitGas, abandon verify and execute operation for temporary
		totalEstimatedFee, err := s.zkbnbClient.EstimateVerifyAndExecuteWithNonce(pendingVerifyAndExecuteBlocks, proofs, gasPrice, 0, nonce)
		if err != nil {
			logx.WithContext(ctx).Errorf("abandon verify block to l1, EstimateGas operation get some error:%s", err.Error())
			return nil, false, err
		}

		verifyAndExecuteBlockData.nonce = nonce
		verifyAndExecuteBlockData.blocks = blocks
		verifyAndExecuteBlockData.proofs = proofs
		verifyAndExecuteBlockData.gasPrice = gasPrice
		verifyAndExecuteBlockData.maxGasPrice = maxGasPrice
		verifyAndExecuteBlockData.totalEstimatedFee = totalEstimatedFee
		verifyAndExecuteBlockData.verifyAndExecuteBlocksInfo = pendingVerifyAndExecuteBlocks

		if !sconfig.GetSenderConfig().VerifyControlSwitch {
			return verifyAndExecuteBlockData, true, nil
		}

		totalTxCount = s.CalculateTotalTxCountForBlock(blocks)
		if totalTxCount <= verifyTxCountLimit && totalEstimatedFee <= maxVerifyTotalGasFee {
			shouldCommit := maxVerifyBlockCount != sconfig.GetSenderConfig().MaxVerifyBlockCount
			logx.WithContext(ctx).Infof("PrepareVerifyAndExecuteBlockData start height=%d,end height=%d,shouldCommit=%s,totalTxCount=%d,verifyTxCountLimit=%d,totalEstimatedFee=%d,maxVerifyTotalGasFee=%d", start, start+int64(maxVerifyBlockCount), shouldCommit, totalTxCount, verifyTxCountLimit, totalEstimatedFee, maxVerifyTotalGasFee)
			return verifyAndExecuteBlockData, shouldCommit, nil
		}
		if maxVerifyBlockCount > 1 {
			maxVerifyBlockCount--
		} else {
			return nil, false, nil
		}
	}
}

func (s *Sender) CheckBlockAndProofData(blocks []*block.Block, blockProofs []*proof.Proof) error {
	if len(blockProofs) != len(blocks) {
		return types.AppErrRelatedProofsNotReady
	}
	// add sanity check
	for i := range blockProofs {
		if blockProofs[i].BlockNumber != blocks[i].BlockHeight {
			return types.AppErrProofNumberNotMatch
		}
	}
	return nil
}

func (s *Sender) PrepareProofData(blockProofs []*proof.Proof) ([]*big.Int, error) {
	var proofs []*big.Int
	for _, bProof := range blockProofs {
		var proofInfo *prove.FormattedProof
		if err := json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo); err != nil {
			return nil, err
		}

		proofs = append(proofs, proofInfo.A[:]...)
		proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
		proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
		proofs = append(proofs, proofInfo.C[:]...)
	}
	return proofs, nil
}

func (s *Sender) PrepareVerifyNonceValue() (uint64, error) {
	nonce, err := s.getProviderClient().GetPendingNonce(s.GetVerifyAddress().Hex())
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce for verify block, errL %v", err)
	}
	return nonce, nil
}

func (s *Sender) PrepareVerifyGasPriceData(ctx context.Context) (*big.Int, decimal.Decimal, error) {
	var err error
	var gasPrice *big.Int
	var emptyMaxGasPrice = decimal.NewFromInt(0)

	if s.config.ChainConfig.GasPrice > 0 {
		gasPrice = big.NewInt(int64(s.config.ChainConfig.GasPrice))
		s.ValidOverSuggestGasPrice(gasPrice)
	} else {
		gasPrice, err = s.getProviderClient().SuggestGasPrice(context.Background())
		if err != nil {
			logx.WithContext(ctx).Errorf("failed to fetch gas price: %v", err)
			return nil, emptyMaxGasPrice, err
		}
	}
	maxGasPrice := (decimal.NewFromInt(gasPrice.Int64()).Mul(decimal.
		NewFromInt(int64(s.config.ChainConfig.MaxGasPriceIncreasePercentage))).
		Div(decimal.NewFromInt(100))).Add(decimal.NewFromInt(gasPrice.Int64()))

	nonce, err := s.getProviderClient().GetPendingNonce(s.GetVerifyAddress().Hex())
	if err != nil {
		return nil, emptyMaxGasPrice, fmt.Errorf("failed to get nonce for verify block, errL %v", err)
	}

	l1RollupTx, err := s.l1RollupTxModel.GetLatestByNonce(int64(nonce), l1rolluptx.TxTypeVerifyAndExecute)
	if err != nil && err != types.DbErrNotFound {
		return nil, emptyMaxGasPrice, fmt.Errorf("failed to get latest l1 rollup tx by nonce %d, err: %v", nonce, err)
	}
	if l1RollupTx != nil && l1RollupTx.L1Nonce == int64(nonce) {
		standByGasPrice := decimal.NewFromInt(l1RollupTx.GasPrice).Add(decimal.NewFromInt(l1RollupTx.GasPrice).Mul(decimal.NewFromFloat(0.1)))
		if standByGasPrice.GreaterThan(maxGasPrice) {
			logx.WithContext(ctx).Errorf("abandon verify block to l1, gasPrice>maxGasPrice,l1 nonce: %d,gasPrice: %d,maxGasPrice: %d", nonce, standByGasPrice, maxGasPrice)
			return nil, emptyMaxGasPrice, nil
		}
		gasPrice = standByGasPrice.RoundUp(0).BigInt()
		logx.WithContext(ctx).Infof("speed up verify block to l1,l1 nonce: %d,gasPrice: %d", nonce, gasPrice)
	}
	return gasPrice, maxGasPrice, nil
}

func (s *Sender) CalculateBlockIntervalForCompressedBlock(blocks []*compressedblock.CompressedBlock) int64 {
	if len(blocks) > 0 {
		block := blocks[0]
		interval := time.Now().Unix() - block.CreatedAt.Unix()
		return interval
	}
	return 0
}

func (s *Sender) CalculateBlockIntervalForBlock(blocks []*block.Block) int64 {
	if len(blocks) > 0 {
		block := blocks[0]
		interval := time.Now().Unix() - block.CreatedAt.Unix()
		return interval
	}
	return 0
}

func (s *Sender) CalculateTotalTxCountForCompressBlock(blocks []*compressedblock.CompressedBlock) uint64 {
	var totalTxCount uint16 = 0
	if len(blocks) > 0 {
		for _, b := range blocks {
			totalTxCount = totalTxCount + b.BlockSize
		}
	}
	return uint64(totalTxCount)
}

func (s *Sender) CalculateTotalTxCountForBlock(blocks []*block.Block) uint64 {
	var totalTxCount uint16 = 0
	if len(blocks) > 0 {
		for _, b := range blocks {
			totalTxCount = totalTxCount + b.BlockSize
		}
		return uint64(totalTxCount)
	}
	return 0
}

func (s *Sender) GenerateConstructorForCommit() (zkbnb.TransactOptsConstructor, error) {
	if s.commitKmsKeyClient != nil {
		return s.commitKmsKeyClient, nil
	} else if s.commitAuthClient != nil {
		return s.commitAuthClient, nil
	}
	return nil, errors.New("both commitKmsKeyClient and commitAuthClient are all not initiated yet")
}

func (s *Sender) IsOverMaxErrorRetryCount(height int64, txType uint) bool {
	cacheKey := fmt.Sprintf("%s-%d-%d", SentBlockToL1ErrorPrefix, txType, height)
	retryCount := 0
	cacheValue, found := s.goCache.Get(cacheKey)
	if found {
		retryCount = cacheValue.(int)
		if retryCount >= MaxErrorRetryCount {
			return true
		}
	}
	return false
}

func (s *Sender) HandleSendTxToL1Error(height int64, txType uint, txHash string, err error) {
	if _, ok := err.(*url.Error); ok {
		return
	}

	ttl := time.Minute * 30
	if strings.Contains(err.Error(), CallContractError) {
		ttl = time.Minute * 120
	}

	cacheKey := fmt.Sprintf("%s-%d-%d", SentBlockToL1ErrorPrefix, txType, height)
	retryCount := 0
	cacheValue, found := s.goCache.Get(cacheKey)
	if found {
		retryCount = cacheValue.(int) + 1
	} else {
		retryCount = 1
	}
	s.goCache.SetWithTTL(cacheKey, retryCount, 0, ttl)
	logx.Infof("fail to send tx to L1, Send tx to L1 has been called %d times, txHash=%s,L2BlockHeight=%d,txType=%d,err=%v", retryCount, txHash, height, txType, err.Error())
}

func (s *Sender) GenerateConstructorForVerifyAndExecute() (zkbnb.TransactOptsConstructor, error) {
	if s.verifyKmsKeyClient != nil {
		return s.verifyKmsKeyClient, nil
	} else if s.verifyAuthClient != nil {
		return s.verifyAuthClient, nil
	}
	return nil, errors.New("both verifyKmsKeyClient and verifyAuthClient are all not initiated yet")
}

func (s *Sender) GetCommitAddress() common.Address {
	if s.commitKmsKeyClient != nil {
		return s.commitKmsKeyClient.GetL1Address()
	} else if s.commitAuthClient != nil {
		return s.commitAuthClient.GetL1Address()
	}
	return [20]byte{}
}

func (s *Sender) GetVerifyAddress() common.Address {
	if s.verifyKmsKeyClient != nil {
		return s.verifyKmsKeyClient.GetL1Address()
	} else if s.verifyAuthClient != nil {
		return s.verifyAuthClient.GetL1Address()
	}
	return [20]byte{}
}

func (s *Sender) Shutdown() {
	sqlDB, err := s.db.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}
}

func (s *Sender) Monitor() {
	//balance
	info, err := s.sysConfigModel.GetSysConfigByName("ZkBNBContract")
	if err == nil {
		s.SetBalance(info, "ZkBNBContract")
	}

	info, err = s.sysConfigModel.GetSysConfigByName("CommitAddress")
	if err == nil {
		s.SetBalance(info, "CommitAddress")
	}

	info, err = s.sysConfigModel.GetSysConfigByName("VerifyAddress")
	if err == nil {
		s.SetBalance(info, "VerifyAddress")
	}
	contractBalanceMetric.WithLabelValues("gasAccount").Set(common2.GetFeeFromWei(s.GetAccount(types.GasAccount)))
	contractBalanceMetric.WithLabelValues("protocolAccount").Set(common2.GetFeeFromWei(s.GetAccount(types.ProtocolAccount)))

	//costs
	s.MetricBatchCommitContact()
	s.MetricBatchVerifyContact()
}

func (s *Sender) GetAccount(accountIndex int64) *big.Int {
	dbAccount, err := s.accountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		return big.NewInt(0)
	}
	var assetInfo map[int64]*types.AccountAsset
	err = json.Unmarshal([]byte(dbAccount.AssetInfo), &assetInfo)
	if err != nil {
		return big.NewInt(0)
	}
	asset, ok := assetInfo[0]
	if ok {
		return asset.Balance
	} else {
		return big.NewInt(0)
	}
}

func (s *Sender) MetricBatchCommitContact() {
	blockHeights := make([]int64, 0)
	txs, err := s.l1RollupTxModel.GetRecent2Transact(l1rolluptx.TxTypeCommit)
	if err == nil {
		if len(txs) == 2 {
			txRollUpStart := txs[1]
			txRollUpEnd := txs[0]
			value := txRollUpEnd.CreatedAt.Unix() - txRollUpStart.CreatedAt.Unix()
			runTimeIntervalMetric.WithLabelValues("commit").Set(float64(value))
		}
	}
	if s.metricCommitRollupTxId == 0 {
		if err == nil {
			if len(txs) == 1 {
				txRollUp := txs[0]
				for i := s.metricCommitRollupHeight + 1; i <= txRollUp.L2BlockHeight; i++ {
					blockHeights = append(blockHeights, i)
				}
				s.setBatchCommitContactMetric(txRollUp, blockHeights)
			} else if len(txs) == 2 {
				//id desc
				txRollUpStart := txs[1]
				txRollUpEnd := txs[0]
				for i := txRollUpStart.L2BlockHeight + 1; i <= txRollUpEnd.L2BlockHeight; i++ {
					blockHeights = append(blockHeights, i)
				}
				s.setBatchCommitContactMetric(txRollUpEnd, blockHeights)
			}
		}
	} else {
		txRollUp, err := s.l1RollupTxModel.GetRecentById(s.metricCommitRollupTxId, l1rolluptx.TxTypeCommit)
		if err == nil {
			for i := s.metricCommitRollupHeight + 1; i <= txRollUp.L2BlockHeight; i++ {
				blockHeights = append(blockHeights, i)
			}
			s.setBatchCommitContactMetric(txRollUp, blockHeights)
		}
	}
}

func (s *Sender) MetricBatchVerifyContact() {
	blockHeights := make([]int64, 0)
	txs, err := s.l1RollupTxModel.GetRecent2Transact(l1rolluptx.TxTypeVerifyAndExecute)
	if err == nil {
		if len(txs) == 2 {
			txRollUpStart := txs[1]
			txRollUpEnd := txs[0]
			value := txRollUpEnd.CreatedAt.Unix() - txRollUpStart.CreatedAt.Unix()
			runTimeIntervalMetric.WithLabelValues("verify").Set(float64(value))
		}
	}
	if s.metricVerifyRollupTxId == 0 {
		if err == nil {
			if len(txs) == 1 {
				txRollUp := txs[0]
				for i := s.metricVerifyRollupHeight + 1; i <= txRollUp.L2BlockHeight; i++ {
					blockHeights = append(blockHeights, i)
				}
				s.setBatchVerifyContactMetric(txRollUp, blockHeights)
			} else if len(txs) == 2 {
				//id desc
				txRollUpStart := txs[1]
				txRollUpEnd := txs[0]
				for i := txRollUpStart.L2BlockHeight + 1; i <= txRollUpEnd.L2BlockHeight; i++ {
					blockHeights = append(blockHeights, i)
				}
				s.setBatchVerifyContactMetric(txRollUpEnd, blockHeights)
			}
		}
	} else {
		txRollUp, err := s.l1RollupTxModel.GetRecentById(s.metricVerifyRollupTxId, l1rolluptx.TxTypeVerifyAndExecute)
		if err == nil {
			for i := s.metricVerifyRollupHeight + 1; i <= txRollUp.L2BlockHeight; i++ {
				blockHeights = append(blockHeights, i)
			}
			s.setBatchVerifyContactMetric(txRollUp, blockHeights)
		}
	}
}

func (s *Sender) setBatchCommitContactMetric(txRollUp *l1rolluptx.L1RollupTx, blockHeights []int64) {
	if txRollUp.GasPrice == int64(0) || txRollUp.GasUsed == 0 {
		return
	}
	gasCost := ffmath.Multiply(new(big.Int).SetUint64(txRollUp.GasUsed), new(big.Int).SetInt64(txRollUp.GasPrice))
	cost := common2.GetFeeFromWei(gasCost)
	batchCommitCostMetric.Add(cost)
	batchTotalCostMetric.Add(cost)
	batchCommitContactMetric.WithLabelValues("gasCost").Set(cost)
	batchCommitContactMetric.WithLabelValues("blockHeight").Set(float64(txRollUp.L2BlockHeight))
	batchCommitContactMetric.WithLabelValues("blockNumber").Set(float64(len(blockHeights)))
	count, err := s.txModel.GetCountByHeights(blockHeights)
	if err == nil {
		batchCommitContactMetric.WithLabelValues("txNumber").Set(float64(count))
		if count == 0 {
			batchCommitContactMetric.WithLabelValues("averageTxCost").Set(0)

		} else {
			average := ffmath.Div(gasCost, big.NewInt(count))
			batchCommitContactMetric.WithLabelValues("averageTxCost").Set(common2.GetFeeFromWei(average))
		}
	} else {
		batchCommitContactMetric.WithLabelValues("txNumber").Set(0)
		batchCommitContactMetric.WithLabelValues("averageTxCost").Set(0)
	}
	s.metricCommitRollupTxId = txRollUp.ID
	s.metricCommitRollupHeight = txRollUp.L2BlockHeight

}

func (s *Sender) setBatchVerifyContactMetric(txRollUp *l1rolluptx.L1RollupTx, blockHeights []int64) {
	if txRollUp.GasPrice == int64(0) || txRollUp.GasUsed == 0 {
		return
	}
	gasCost := ffmath.Multiply(new(big.Int).SetUint64(txRollUp.GasUsed), new(big.Int).SetInt64(txRollUp.GasPrice))
	cost := common2.GetFeeFromWei(gasCost)
	batchVerifyCostMetric.Add(cost)
	batchTotalCostMetric.Add(cost)
	batchVerifyContactMetric.WithLabelValues("gasCost").Set(cost)
	batchVerifyContactMetric.WithLabelValues("blockHeight").Set(float64(txRollUp.L2BlockHeight))
	batchVerifyContactMetric.WithLabelValues("blockNumber").Set(float64(len(blockHeights)))
	count, err := s.txModel.GetCountByHeights(blockHeights)
	if err == nil {
		batchVerifyContactMetric.WithLabelValues("txNumber").Set(float64(count))
		if count == 0 {
			batchVerifyContactMetric.WithLabelValues("averageTxCost").Set(0)

		} else {
			average := ffmath.Div(gasCost, big.NewInt(count))
			batchVerifyContactMetric.WithLabelValues("averageTxCost").Set(common2.GetFeeFromWei(average))
		}
	} else {
		batchVerifyContactMetric.WithLabelValues("txNumber").Set(0)
		batchVerifyContactMetric.WithLabelValues("averageTxCost").Set(0)
	}
	s.metricVerifyRollupTxId = txRollUp.ID
	s.metricVerifyRollupHeight = txRollUp.L2BlockHeight
}

func (s *Sender) SetBalance(info *sysconfig.SysConfig, name string) {
	balance, err := s.getProviderClient().GetBalance(info.Value)
	if err != nil {
		contractBalanceMetric.WithLabelValues(name).Set(float64(0))
		logx.Errorf("%s get balance error: %s", name, err.Error())
	}
	contractBalanceMetric.WithLabelValues(name).Set(common2.GetFeeFromWei(balance))
}

func (s *Sender) getZkbnbClient(cli *rpc.ProviderClient) *zkbnb.ZkBNBClient {
	commitConstructor, err := s.GenerateConstructorForCommit()
	if err != nil {
		logx.Severef("fatal error, GenerateConstructorForCommit raises error:%v", err)
		panic("fatal error, GenerateConstructorForCommit raises error:" + err.Error())
	}
	verifyConstructor, err := s.GenerateConstructorForVerifyAndExecute()
	if err != nil {
		logx.Severef("fatal error, GenerateConstructorForVerifyAndExecute raises error:%v", err)
		panic("fatal error, GenerateConstructorForVerifyAndExecute raises error:" + err.Error())
	}
	zkbnbClient, err := zkbnb.NewZkBNBClient(cli, s.ZkBNBContractAddress)
	if err != nil {
		logx.Severef("fatal error, ZkBNBClient initiate raises error:%v", err)
		panic("fatal error, ZkBNBClient initiate raises error:" + err.Error())
	}
	zkbnbClient.CommitConstructor = commitConstructor
	zkbnbClient.VerifyConstructor = verifyConstructor
	return zkbnbClient
}

func (s *Sender) getProviderClient() *rpc.ProviderClient {
	return rpc_client.GetRpcClient()
}

func (s *Sender) ValidOverSuggestGasPrice150Percent(gasPrice *big.Int) {
	suggestGasPrice, err := s.getProviderClient().SuggestGasPrice(context.Background())
	if err == nil {
		suggestGasPriceMax := ffmath.Add(suggestGasPrice, ffmath.Div(suggestGasPrice, big.NewInt(2)))
		if gasPrice.Cmp(suggestGasPriceMax) > 0 {
			logx.Severef("More than 150% of the suggest gas price,suggestGasPrice=%s,gasPrice=%s", suggestGasPrice.String(), gasPrice.String())
		}
	}
}

func (s *Sender) ValidOverSuggestGasPrice(gasPrice *big.Int) {
	suggestGasPrice, err := s.getProviderClient().SuggestGasPrice(context.Background())
	if err == nil {
		if gasPrice.Cmp(suggestGasPrice) > 0 {
			logx.Severef("The gasPrice of the apollo configuration is more than the suggest gas price,suggestGasPrice=%s,gasPrice=%s", suggestGasPrice.String(), gasPrice.String())
		}
	}
}

func (s *Sender) TimeOut() {
	maxCommitBlockInterval := sconfig.GetSenderConfig().MaxCommitBlockInterval
	maxCommitBlockTime, _ := time.ParseDuration(fmt.Sprintf("-%ds", maxCommitBlockInterval))
	commitTime := time.Now().Add(maxCommitBlockTime)
	commitBlock, err := s.blockModel.GetBlockByStatusAndTime(tx.StatusPacked, commitTime)
	if err == nil {
		interval := time.Now().Unix() - commitBlock.CreatedAt.Unix()
		runTimeIntervalMetric.WithLabelValues("commitBlockTimeOut").Set(float64(interval))
	} else {
		runTimeIntervalMetric.WithLabelValues("commitBlockTimeOut").Set(float64(0))
	}
	maxVerifyBlockInterval := sconfig.GetSenderConfig().MaxVerifyBlockInterval
	maxVerifyBlockTime, _ := time.ParseDuration(fmt.Sprintf("-%ds", maxVerifyBlockInterval))
	verifyTime := time.Now().Add(maxVerifyBlockTime)
	verifyBlock, err := s.blockModel.GetBlockByStatusAndTime(tx.StatusCommitted, verifyTime)
	if err == nil {
		interval := time.Now().Unix() - verifyBlock.CreatedAt.Unix()
		runTimeIntervalMetric.WithLabelValues("verifyBlockTimeOut").Set(float64(interval))
	} else {
		runTimeIntervalMetric.WithLabelValues("verifyBlockTimeOut").Set(float64(0))
	}

}
