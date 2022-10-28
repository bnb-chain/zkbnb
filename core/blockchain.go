package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

const (
	defaultTaskPoolSize = 1000
)

type ChainConfig struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	//nolint:staticcheck
	CacheConfig statedb.CacheConfig `json:",optional"`
	TreeDB      struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption      tree.RedisDBOption `json:",optional"`
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
	taskPool     *ants.Pool
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	db, err := gorm.Open(postgres.Open(config.Postgres.DataSource))
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	bc := &BlockChain{
		ChainDB:     sdb.NewChainDB(db),
		chainConfig: config,
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
	if bc.currentBlock.BlockStatus == block.StatusProposing {
		curHeight--
	}
	redisCache := dbcache.NewRedisCache(config.CacheRedis[0].Host, config.CacheRedis[0].Pass, 15*time.Minute)
	treeCtx := &tree.Context{
		Name:          moduleName,
		Driver:        config.TreeDB.Driver,
		LevelDBOption: &config.TreeDB.LevelDBOption,
		RedisDBOption: &config.TreeDB.RedisDBOption,
	}

	bc.Statedb, err = sdb.NewStateDB(treeCtx, bc.ChainDB, redisCache, &config.CacheConfig, config.TreeDB.AssetTreeCacheSize, bc.currentBlock.StateRoot, curHeight)
	if err != nil {
		return nil, err
	}
	bc.processor = NewCommitProcessor(bc)
	taskPool, err := ants.NewPool(defaultTaskPoolSize)
	if err != nil {
		return nil, err
	}
	bc.taskPool = taskPool

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

func (bc *BlockChain) CommitNewBlock(blockSize int, createdAt int64) (*block.BlockStates, error) {
	newBlock, compressedBlock, err := bc.commitNewBlock(blockSize, createdAt)
	if err != nil {
		return nil, err
	}

	currentHeight := bc.currentBlock.BlockHeight

	err = tree.CommitTrees(bc.taskPool, uint64(currentHeight), bc.Statedb.AccountTree, bc.Statedb.AccountAssetTrees, bc.Statedb.NftTree)
	if err != nil {
		return nil, err
	}

	pendingAccount, pendingAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight)
	if err != nil {
		return nil, err
	}

	pendingNft, pendingNftHistory, err := bc.Statedb.GetPendingNft(currentHeight)
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

func (bc *BlockChain) commitNewBlock(blockSize int, createdAt int64) (*block.Block, *compressedblock.CompressedBlock, error) {
	s := bc.Statedb
	if blockSize < len(s.Txs) {
		return nil, nil, errors.New("block size too small")
	}

	newBlock := bc.currentBlock
	if newBlock.BlockStatus != block.StatusProposing {
		newBlock = &block.Block{
			Model: gorm.Model{
				CreatedAt: time.UnixMilli(createdAt),
			},
			BlockHeight: bc.currentBlock.BlockHeight + 1,
			StateRoot:   bc.currentBlock.StateRoot,
			BlockStatus: block.StatusProposing,
		}
	}

	// Intermediate state root.
	err := s.IntermediateRoot(false)
	if err != nil {
		return nil, nil, err
	}

	// Align pub data.
	s.AlignPubData(blockSize)

	commitment := chain.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(newBlock.StateRoot), common.FromHex(s.StateRoot),
		s.PubData, int64(len(s.PubDataOffset)))

	newBlock.BlockSize = uint16(blockSize)
	newBlock.BlockCommitment = commitment
	newBlock.StateRoot = s.StateRoot
	newBlock.PriorityOperations = s.PriorityOperations
	newBlock.PendingOnChainOperationsHash = common.Bytes2Hex(s.PendingOnChainOperationsHash)
	newBlock.Txs = s.Txs
	for _, executedTx := range newBlock.Txs {
		executedTx.TxStatus = tx.StatusPacked
	}
	newBlock.BlockStatus = block.StatusPending
	if len(s.PendingOnChainOperationsPubData) > 0 {
		onChainOperationsPubDataBytes, err := json.Marshal(s.PendingOnChainOperationsPubData)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(s.PubDataOffset)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}
	newCompressedBlock := &compressedblock.CompressedBlock{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(s.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}

	bc.currentBlock = newBlock
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
