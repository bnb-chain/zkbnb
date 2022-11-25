package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/zkbnbprometheus"
	"github.com/bnb-chain/zkbnb/core/statedb"
	sdb "github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

// metrics
var (
	updateAccountTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_asset_smt",
		Help:      "update asset smt tree operation time",
	})

	commitAccountTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_smt",
		Help:      "commit account smt tree operation time",
	})

	executeTxPrepareMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_prepare_time",
		Help:      "execute txs prepare operation time",
	})

	executeTxVerifyInputsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_verify_inputs_time",
		Help:      "execute txs verify inputs operation time",
	})

	executeGenerateTxDetailsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_tx_details_time",
		Help:      "execute txs generate tx details operation time",
	})

	executeTxApplyTransactionMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_apply_transaction_time",
		Help:      "execute txs apply transaction operation time",
	})

	executeTxGeneratePubDataMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_pub_data_time",
		Help:      "execute txs generate pub data operation time",
	})
	executeTxGetExecutedTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_get_executed_tx_time",
		Help:      "execute txs get executed tx operation time",
	})
)

type ChainConfig struct {
	Postgres struct {
		DataSource string
		LogLevel   logger.LogLevel `json:",optional"`
	}
	CacheRedis cache.CacheConf
	//nolint:staticcheck
	CacheConfig statedb.CacheConfig `json:",optional"`
	TreeDB      struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
	}
}

type BlockChain struct {
	*sdb.ChainDB
	Statedb *sdb.StateDB // Cache for current block changes.

	chainConfig *ChainConfig
	dryRun      bool //dryRun mode is used for verifying user inputs, is not for execution

	currentBlock *block.Block
	processor    Processor
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	db, err := gorm.Open(postgres.Open(config.Postgres.DataSource), &gorm.Config{
		Logger: logger.Default.LogMode(config.Postgres.LogLevel),
	})
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	bc := &BlockChain{
		ChainDB:     sdb.NewChainDB(db),
		chainConfig: config,
	}

	err = bc.TxPoolModel.UpdateTxsToPending()
	if err != nil {
		logx.Error("update pool tx to pending failed: ", err)
		panic("update pool tx to pending failed: " + err.Error())
	}

	blockHeights, err := bc.BlockModel.GetProposingBlockHeights()
	if err != nil {
		logx.Error("get proposing block height failed: ", err)
		panic("delete block failed: " + err.Error())
	}
	if blockHeights != nil {
		logx.Infof("get proposing block heights: %v", blockHeights)
		err = bc.BlockModel.DeleteProposingBlock()
		if err != nil {
			logx.Error("delete block failed: ", err)
			panic("delete block failed: " + err.Error())
		}
	}
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block failed: ", err)
		return nil, err
	}
	logx.Infof("get current block height: %d", curHeight)

	bc.currentBlock, err = bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}
	if bc.currentBlock.BlockStatus == block.StatusProposing {
		logx.Errorf("current block status is StatusProposing,invalid block, height=%d", bc.currentBlock.BlockHeight)
		panic("current block status is StatusProposing,invalid block, height=" + strconv.FormatInt(bc.currentBlock.BlockHeight, 10))
	}
	//todo config

	redisCache := dbcache.NewRedisCache(config.CacheRedis[0].Host, config.CacheRedis[0].Pass, 15*time.Minute)
	treeCtx, err := tree.NewContext(moduleName, config.TreeDB.Driver, false, config.TreeDB.RoutinePoolSize, &config.TreeDB.LevelDBOption, &config.TreeDB.RedisDBOption)
	if err != nil {
		return nil, err
	}

	treeCtx.SetOptions(bsmt.BatchSizeLimit(3 * 1024 * 1024))
	bc.Statedb, err = sdb.NewStateDB(treeCtx, bc.ChainDB, redisCache, &config.CacheConfig, config.TreeDB.AssetTreeCacheSize, bc.currentBlock.StateRoot, curHeight)
	if err != nil {
		return nil, err
	}
	bc.Statedb.PreviousStateRoot = bc.currentBlock.StateRoot

	accountFromDbGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_from_db_time",
		Help:      "account from db time",
	})

	accountGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_time",
		Help:      "account time",
	})

	verifyGasGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifyGasGauge_time",
		Help:      "verifyGas time",
	})
	verifySignature := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifySignature_time",
		Help:      "verifySignature time",
	})

	accountTreeMultiSetGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "accountTreeMultiSetGauge_time",
		Help:      "accountTreeMultiSetGauge time",
	})

	if err := prometheus.Register(verifyGasGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register verifyGasGauge error: %v", err)
	}

	if err := prometheus.Register(verifySignature); err != nil {
		return nil, fmt.Errorf("prometheus.Register verifySignature error: %v", err)
	}

	if err := prometheus.Register(accountTreeMultiSetGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountTreeMultiSetGauge error: %v", err)
	}

	if err := prometheus.Register(accountFromDbGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountFromDbMetrics error: %v", err)
	}
	getAccountCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_counter",
		Help:      "get account counter",
	})
	if err := prometheus.Register(getAccountCounter); err != nil {
		return nil, fmt.Errorf("prometheus.Register getAccountCounter error: %v", err)
	}

	getAccountFromDbCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_from_db_counter",
		Help:      "get account from db counter",
	})
	if err := prometheus.Register(getAccountFromDbCounter); err != nil {
		return nil, fmt.Errorf("prometheus.Register getAccountFromDbCounter error: %v", err)
	}

	accountTreeTimeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_account_tree_time",
			Help:      "get_account_tree_time.",
		},
		[]string{"type"})
	if err := prometheus.Register(accountTreeTimeGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register accountTreeTimeGauge error: %v", err)
	}

	nftTreeTimeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_nft_tree_time",
			Help:      "get_nft_tree_time.",
		},
		[]string{"type"})
	if err := prometheus.Register(nftTreeTimeGauge); err != nil {
		return nil, fmt.Errorf("prometheus.Register nftTreeTimeGauge error: %v", err)
	}

	stateDBMetrics := &zkbnbprometheus.StateDBMetrics{
		GetAccountFromDbGauge:    accountFromDbGauge,
		GetAccountGauge:          accountGauge,
		GetAccountCounter:        getAccountCounter,
		GetAccountFromDbCounter:  getAccountFromDbCounter,
		VerifyGasGauge:           verifyGasGauge,
		VerifySignature:          verifySignature,
		AccountTreeGauge:         accountTreeTimeGauge,
		NftTreeGauge:             nftTreeTimeGauge,
		AccountTreeMultiSetGauge: accountTreeMultiSetGauge,
	}
	bc.Statedb.Metrics = stateDBMetrics

	if err := prometheus.Register(executeTxPrepareMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxPrepareMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxVerifyInputsMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxVerifyInputsMetrics error: %v", err)
	}

	if err := prometheus.Register(executeGenerateTxDetailsMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeGenerateTxDetailsMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxApplyTransactionMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxApplyTransactionMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxGeneratePubDataMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxGeneratePubDataMetrics error: %v", err)
	}

	if err := prometheus.Register(executeTxGetExecutedTxMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxGetExecutedTxMetrics error: %v", err)
	}
	prometheusMetrics := &zkbnbprometheus.Metrics{
		TxPrepareMetrics:           executeTxPrepareMetrics,
		TxVerifyInputsMetrics:      executeTxVerifyInputsMetrics,
		TxGenerateTxDetailsMetrics: executeGenerateTxDetailsMetrics,
		TxApplyTransactionMetrics:  executeTxApplyTransactionMetrics,
		TxGeneratePubDataMetrics:   executeTxGeneratePubDataMetrics,
		TxGetExecutedTxMetrics:     executeTxGetExecutedTxMetrics,
	}
	bc.processor = NewCommitProcessor(bc, prometheusMetrics)

	// register metrics
	if err := prometheus.Register(updateAccountTreeMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updateAccountTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(commitAccountTreeMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register commitAccountTreeMetrics error: %v", err)
	}

	return bc, nil
}

// NewBlockChainForDryRun - for dry run mode, we can reuse existing models for quick creation
// , e.g., for sending tx, we can create blockchain for each request quickly
func NewBlockChainForDryRun(accountModel account.AccountModel,
	nftModel nft.L2NftModel, txPoolModel tx.TxPoolModel, assetModel asset.AssetModel,
	sysConfigModel sysconfig.SysConfigModel, redisCache dbcache.Cache) (*BlockChain, error) {
	chainDb := &sdb.ChainDB{
		AccountModel:     accountModel,
		L2NftModel:       nftModel,
		TxPoolModel:      txPoolModel,
		L2AssetInfoModel: assetModel,
		SysConfigModel:   sysConfigModel,
	}
	statedb, err := sdb.NewStateDBForDryRun(redisCache, &statedb.DefaultCacheConfig, chainDb)
	if err != nil {
		return nil, err
	}
	bc := &BlockChain{
		ChainDB: chainDb,
		dryRun:  true,
		Statedb: statedb,
	}
	bc.processor = NewAPIProcessor(bc)
	return bc, nil
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx) error {
	return bc.processor.Process(tx)
}

func (bc *BlockChain) InitNewBlock() (*block.Block, error) {
	newBlock := &block.Block{
		Model: gorm.Model{
			// The block timestamp will be set when the first transaction executed.
			CreatedAt: time.Time{},
		},
		BlockHeight: bc.currentBlock.BlockHeight + 1,
		StateRoot:   bc.currentBlock.StateRoot,
		BlockStatus: block.StatusProposing,
	}

	bc.currentBlock = newBlock
	bc.Statedb.PurgeCache(bc.currentBlock.StateRoot)
	err := bc.Statedb.MarkGasAccountAsPending()
	return newBlock, err
}

func (bc *BlockChain) CurrentBlock() *block.Block {
	return bc.currentBlock
}

//func (bc *BlockChain) CommitNewBlock(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.BlockStates, error) {
//	newBlock, compressedBlock, err := bc.commitNewBlock(blockSize, stateDataCopy)
//	if err != nil {
//		return nil, err
//	}
//
//	currentHeight := stateDataCopy.CurrentBlock.BlockHeight
//
//	start := time.Now()
//	err = tree.AccountTreeAndNftTreeMultiSet(uint64(bc.StateDB().GetPrunedBlockHeight()), bc.Statedb.AccountTree, bc.Statedb.AccountAssetTrees, bc.Statedb.NftTree)
//	if err != nil {
//		return nil, err
//	}
//	commitAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
//
//	pendingAccount, pendingAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight, stateDataCopy)
//	if err != nil {
//		return nil, err
//	}
//
//	pendingNft, pendingNftHistory, err := bc.Statedb.GetPendingNft(currentHeight, stateDataCopy)
//	if err != nil {
//		return nil, err
//	}
//
//	return &block.BlockStates{
//		Block:                 newBlock,
//		CompressedBlock:       compressedBlock,
//		PendingAccount:        pendingAccount,
//		PendingAccountHistory: pendingAccountHistory,
//		PendingNft:            pendingNft,
//		PendingNftHistory:     pendingNftHistory,
//	}, nil
//}

func (bc *BlockChain) UpdateAccountAssetTree(stateDataCopy *statedb.StateDataCopy) error {
	start := time.Now()
	// Intermediate state root.
	err := bc.Statedb.IntermediateRoot(false, stateDataCopy)
	if err != nil {
		return err
	}
	updateAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

func (bc *BlockChain) UpdateAccountTreeAndNftTree(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.BlockStates, error) {
	newBlock := stateDataCopy.CurrentBlock
	if newBlock.BlockStatus != block.StatusProposing {
		newBlock = &block.Block{
			Model: gorm.Model{
				CreatedAt: time.UnixMilli(stateDataCopy.CurrentBlock.CreatedAt.UnixMilli()),
			},
			BlockHeight: stateDataCopy.CurrentBlock.BlockHeight + 1,
			StateRoot:   stateDataCopy.CurrentBlock.StateRoot,
			BlockStatus: block.StatusProposing,
		}
	}
	err := bc.Statedb.AccountTreeAndNftTreeMultiSet(stateDataCopy)
	if err != nil {
		return nil, err
	}
	// Align pub data.
	bc.Statedb.AlignPubData(blockSize, stateDataCopy)

	commitment := chain.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(bc.Statedb.PreviousStateRoot), common.FromHex(stateDataCopy.StateCache.StateRoot),
		stateDataCopy.StateCache.PubData, int64(len(stateDataCopy.StateCache.PubDataOffset)))

	newBlock.BlockSize = uint16(blockSize)
	newBlock.BlockCommitment = commitment
	newBlock.StateRoot = stateDataCopy.StateCache.StateRoot
	newBlock.PriorityOperations = stateDataCopy.StateCache.PriorityOperations
	newBlock.PendingOnChainOperationsHash = common.Bytes2Hex(stateDataCopy.StateCache.PendingOnChainOperationsHash)
	newBlock.Txs = stateDataCopy.StateCache.Txs
	for _, executedTx := range newBlock.Txs {
		executedTx.TxStatus = tx.StatusPacked
	}
	newBlock.BlockStatus = block.StatusPending
	if len(stateDataCopy.StateCache.PendingOnChainOperationsPubData) > 0 {
		onChainOperationsPubDataBytes, err := json.Marshal(stateDataCopy.StateCache.PendingOnChainOperationsPubData)
		if err != nil {
			return nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(stateDataCopy.StateCache.PubDataOffset)
	if err != nil {
		return nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}
	newCompressedBlock := &compressedblock.CompressedBlock{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(stateDataCopy.StateCache.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}
	bc.Statedb.PreviousStateRoot = stateDataCopy.StateCache.StateRoot
	//bc.currentBlock = newBlock
	//todo

	currentHeight := stateDataCopy.CurrentBlock.BlockHeight

	start := time.Now()
	//asset.LatestVersion()
	//uint64(bc.StateDB().GetPrunedBlockHeight())
	logx.Infof("CommitAccountTreeAndNftTree,latestVersion=%d,prunedBlockHeight=%d", uint64(bc.Statedb.AccountTree.LatestVersion()), uint64(bc.StateDB().GetPrunedBlockHeight()))
	err = tree.CommitAccountTreeAndNftTree(uint64(bc.Statedb.AccountTree.LatestVersion())-10, bc.Statedb.AccountTree, bc.Statedb.AccountAssetTrees, bc.Statedb.NftTree)
	if err != nil {
		return nil, err
	}
	commitAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

	pendingAccount, pendingAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}

	pendingNft, pendingNftHistory, err := bc.Statedb.GetPendingNft(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}
	return &block.BlockStates{
		Block:                 newBlock,
		CompressedBlock:       newCompressedBlock,
		PendingAccount:        pendingAccount,
		PendingAccountHistory: pendingAccountHistory,
		PendingNft:            pendingNft,
		PendingNftHistory:     pendingNftHistory,
	}, nil
}

func (bc *BlockChain) VerifyExpiredAt(expiredAt int64) error {
	if !bc.dryRun {
		if expiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	} else {
		if expiredAt < time.Now().UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	}
	return nil
}

func (bc *BlockChain) VerifyNonce(accountIndex int64, nonce int64) error {
	if !bc.dryRun {
		expectNonce, err := bc.Statedb.GetCommittedNonce(accountIndex)
		if err != nil {
			return err
		}
		if nonce != expectNonce {
			logx.Infof("committer verify nonce failed,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
			bc.Statedb.ClearPendingNonceFromRedisCache(accountIndex)
			return types.AppErrInvalidNonce
		}
	} else {
		pendingNonce, err := bc.Statedb.GetPendingNonceFromCache(accountIndex)
		if err != nil {
			return err
		}
		if pendingNonce != nonce {
			logx.Infof("clear pending nonce from redis cache,accountIndex=%d,pendingNonce=%d,nonce=%d", accountIndex, pendingNonce, nonce)
			bc.Statedb.ClearPendingNonceFromRedisCache(accountIndex)
			return types.AppErrInvalidNonce
		}
	}
	return nil
}

func (bc *BlockChain) VerifyGas(gasAccountIndex, gasFeeAssetId int64, txType int, gasFeeAmount *big.Int, skipGasAmtChk bool) error {
	cfgGasAccountIndex, err := bc.Statedb.GetGasAccountIndex()
	if err != nil {
		return err
	}
	if gasAccountIndex != cfgGasAccountIndex {
		return types.AppErrInvalidGasFeeAccount
	}

	cfgGasFee, err := bc.Statedb.GetGasConfig()
	if err != nil {
		return err
	}

	gasAsset, ok := cfgGasFee[uint32(gasFeeAssetId)]
	if !ok {
		logx.Errorf("cannot find gas config for asset id: %d", gasFeeAssetId)
		return types.AppErrInvalidGasFeeAsset
	}

	if !skipGasAmtChk {
		gasFee, ok := gasAsset[txType]
		if !ok {
			return errors.New("invalid tx type")
		}
		if gasFeeAmount.Cmp(big.NewInt(gasFee)) < 0 {
			return types.AppErrInvalidGasFeeAmount
		}
	}
	return nil
}

func (bc *BlockChain) StateDB() *sdb.StateDB {
	return bc.Statedb
}

func (bc *BlockChain) DB() *sdb.ChainDB {
	return bc.ChainDB
}

func (bc *BlockChain) setCurrentBlockTimeStamp() {
	if bc.currentBlock.CreatedAt.IsZero() && len(bc.Statedb.Txs) == 0 {
		creatAt := time.Now().UnixMilli()
		bc.currentBlock.CreatedAt = time.UnixMilli(creatAt)
	}
}

func (bc *BlockChain) resetCurrentBlockTimeStamp() {
	if len(bc.Statedb.Txs) > 0 {
		return
	}

	bc.currentBlock.CreatedAt = time.Time{}
}
