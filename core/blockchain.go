package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm/logger"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
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
	updateTreeMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_smt",
		Help:      "update smt tree operation time",
	})

	commitTreeMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_smt",
		Help:      "commit smt tree operation time",
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
		panic("update pool tx to pending failed: " + err.Error())
	}
	err = bc.BlockModel.DeleteProposingBlock()
	if err != nil {
		panic("delete block failed: " + err.Error())
	}

	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block failed: ", err)
		return nil, err
	}

	bc.currentBlock, err = bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}
	//if bc.currentBlock.BlockStatus == block.StatusProposing {
	//	curHeight--
	//}
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
	bc.processor = NewCommitProcessor(bc)

	// register metrics
	if err := prometheus.Register(updateTreeMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updateTreeMetics error: %v", err)
	}
	if err := prometheus.Register(commitTreeMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register commitTreeMetics error: %v", err)
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

func (bc *BlockChain) CommitNewBlock(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.BlockStates, error) {
	newBlock, compressedBlock, err := bc.commitNewBlock(blockSize, stateDataCopy)
	if err != nil {
		return nil, err
	}

	currentHeight := stateDataCopy.CurrentBlock.BlockHeight

	start := time.Now()
	err = tree.CommitTrees(uint64(bc.StateDB().GetPrunedBlockHeight()), bc.Statedb.AccountTree, bc.Statedb.AccountAssetTrees, bc.Statedb.NftTree)
	if err != nil {
		return nil, err
	}
	commitTreeMetics.Set(float64(time.Since(start).Milliseconds()))

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
		CompressedBlock:       compressedBlock,
		PendingAccount:        pendingAccount,
		PendingAccountHistory: pendingAccountHistory,
		PendingNft:            pendingNft,
		PendingNftHistory:     pendingNftHistory,
	}, nil
}

func (bc *BlockChain) commitNewBlock(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.Block, *compressedblock.CompressedBlock, error) {
	if blockSize < len(stateDataCopy.StateCache.Txs) {
		return nil, nil, errors.New("block size too small")
	}

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

	start := time.Now()
	// Intermediate state root.
	err := bc.Statedb.IntermediateRoot(false, stateDataCopy)
	if err != nil {
		return nil, nil, err
	}
	updateTreeMetics.Set(float64(time.Since(start).Milliseconds()))

	// Align pub data.
	bc.Statedb.AlignPubData(blockSize, stateDataCopy)

	commitment := chain.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(newBlock.StateRoot), common.FromHex(stateDataCopy.StateCache.StateRoot),
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
			return nil, nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(stateDataCopy.StateCache.PubDataOffset)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}
	newCompressedBlock := &compressedblock.CompressedBlock{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(stateDataCopy.StateCache.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}

	//bc.currentBlock = newBlock
	//todo
	return newBlock, newCompressedBlock, nil
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
			return types.AppErrInvalidNonce
		}
	} else {
		pendingNonce, err := bc.Statedb.GetPendingNonce(accountIndex)
		if err != nil {
			return err
		}
		if pendingNonce != nonce {
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
