package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/chain"
	sdb "github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/blockforcommit"
	"github.com/bnb-chain/zkbas/dao/dbcache"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

type ChainConfig struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver        tree.Driver
		LevelDBOption tree.LevelDBOption `json:",optional"`
		RedisDBOption tree.RedisDBOption `json:",optional"`
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
	gormPointer, err := gorm.Open(postgres.Open(config.Postgres.DataSource))
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	conn := sqlx.NewSqlConn("postgres", config.Postgres.DataSource)
	bc := &BlockChain{
		ChainDB:     sdb.NewChainDB(conn, config.CacheRedis, gormPointer),
		chainConfig: config,
	}

	curHeight, err := bc.BlockModel.GetCurrentHeight()
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
	bc.Statedb, err = sdb.NewStateDB(treeCtx, bc.ChainDB, redisCache, bc.currentBlock.StateRoot, curHeight)
	if err != nil {
		return nil, err
	}
	bc.processor = NewCommitProcessor(bc)
	return bc, nil
}

// NewBlockChainForDryRun - for dry run mode, we can reuse existing models for quick creation
// , e.g., for sending tx, we can create blockchain for each request quickly
func NewBlockChainForDryRun(accountModel account.AccountModel, liquidityModel liquidity.LiquidityModel,
	nftModel nft.L2NftModel, mempoolModel mempool.MempoolModel, redisCache dbcache.Cache) *BlockChain {
	chainDb := &sdb.ChainDB{
		AccountModel:   accountModel,
		LiquidityModel: liquidityModel,
		L2NftModel:     nftModel,
		MempoolModel:   mempoolModel,
	}
	bc := &BlockChain{
		ChainDB: chainDb,
		dryRun:  true,
		Statedb: sdb.NewStateDBForDryRun(redisCache, chainDb),
	}
	return bc
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx) error {
	return bc.processor.Process(tx)
}

func (bc *BlockChain) ProposeNewBlock() (*block.Block, error) {
	createdAt := time.Now().UnixMilli()
	newBlock := &block.Block{
		Model: gorm.Model{
			CreatedAt: time.UnixMilli(createdAt),
		},
		BlockHeight: bc.currentBlock.BlockHeight + 1,
		StateRoot:   bc.currentBlock.StateRoot,
		BlockStatus: block.StatusProposing,
	}

	err := bc.BlockModel.CreateNewBlock(newBlock)
	if err != nil {
		return nil, err
	}

	bc.currentBlock = newBlock
	bc.Statedb.PurgeCache(bc.currentBlock.StateRoot)
	return newBlock, nil
}

func (bc *BlockChain) CurrentBlock() *block.Block {
	return bc.currentBlock
}

func (bc *BlockChain) CommitNewBlock(blockSize int, createdAt int64) (*block.BlockStates, error) {
	newBlock, newBlockForCommit, err := bc.commitNewBlock(blockSize, createdAt)
	if err != nil {
		return nil, err
	}

	currentHeight := bc.currentBlock.BlockHeight
	pendingNewAccount, pendingUpdateAccount, pendingNewAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight)
	if err != nil {
		return nil, err
	}

	pendingNewLiquidity, pendingUpdateLiquidity, pendingNewLiquidityHistory, err := bc.Statedb.GetPendingLiquidity(currentHeight)
	if err != nil {
		return nil, err
	}

	pendingNewNft, pendingUpdateNft, pendingNewNftHistory, err := bc.Statedb.GetPendingNft(currentHeight)
	if err != nil {
		return nil, err
	}

	return &block.BlockStates{
		Block:                        newBlock,
		BlockForCommit:               newBlockForCommit,
		PendingNewAccount:            pendingNewAccount,
		PendingUpdateAccount:         pendingUpdateAccount,
		PendingNewAccountHistory:     pendingNewAccountHistory,
		PendingNewLiquidity:          pendingNewLiquidity,
		PendingUpdateLiquidity:       pendingUpdateLiquidity,
		PendingNewLiquidityHistory:   pendingNewLiquidityHistory,
		PendingNewNft:                pendingNewNft,
		PendingUpdateNft:             pendingUpdateNft,
		PendingNewNftHistory:         pendingNewNftHistory,
		PendingNewNftWithdrawHistory: bc.Statedb.PendingNewNftWithdrawHistory,
		PendingNewOffer:              bc.Statedb.PendingNewOffer,
		PendingNewL2NftExchange:      bc.Statedb.PendingNewL2NftExchange,
	}, nil
}

func (bc *BlockChain) commitNewBlock(blockSize int, createdAt int64) (*block.Block, *blockforcommit.BlockForCommit, error) {
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
	newBlock.BlockStatus = block.StatusPending
	if s.PendingOnChainOperationsPubData != nil {
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
	newBlockForCommit := &blockforcommit.BlockForCommit{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(s.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}

	bc.currentBlock = newBlock
	return newBlock, newBlockForCommit, nil
}

func (bc *BlockChain) VerifyExpiredAt(expiredAt int64) error {
	if !bc.dryRun {
		if expiredAt != types.NilExpiredAt && expiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
			return errors.New("invalid ExpiredAt")
		}
	} else {
		if expiredAt != types.NilExpiredAt && expiredAt < time.Now().UnixMilli() {
			return errors.New("invalid ExpiredAt")
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
			return errors.New("invalid Nonce")
		}
	} else {
		pendingNonce, err := bc.Statedb.GetPendingNonce(accountIndex)
		if err != nil {
			return err
		}
		if pendingNonce != nonce {
			return errors.New("invalid Nonce")
		}
	}
	return nil
}

func (bc *BlockChain) StateDB() *sdb.StateDB {
	return bc.Statedb
}

func (bc *BlockChain) DB() *sdb.ChainDB {
	return bc.DB()
}
