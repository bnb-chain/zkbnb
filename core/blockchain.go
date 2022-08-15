package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/dbcache"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/treedb"
	"github.com/bnb-chain/zkbas/common/util"
)

var (
	ZeroBigInt = big.NewInt(0)
)

type ChainConfig struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	TreeDB     struct {
		Driver        treedb.Driver
		LevelDBOption treedb.LevelDBOption `json:",optional"`
		RedisDBOption treedb.RedisDBOption `json:",optional"`
	}
}

const (
	_ = iota
	StateCachePending
	StateCacheCached
)

type StateCache struct {
	stateRoot string
	txs       []*tx.Tx

	// Updated in executor's ApplyTransaction method.
	pendingNewAccountIndexMap      map[int64]int
	pendingNewLiquidityIndexMap    map[int64]int
	pendingNewNftIndexMap          map[int64]int
	pendingUpdateAccountIndexMap   map[int64]int
	pendingUpdateLiquidityIndexMap map[int64]int
	pendingUpdateNftIndexMap       map[int64]int
	pendingNewNftWithdrawHistory   []*nft.L2NftWithdrawHistory

	// Updated in executor's GeneratePubData method.
	pubData                         []byte
	priorityOperations              int64
	pubDataOffset                   []uint32
	pendingOnChainOperationsPubData [][]byte
	pendingOnChainOperationsHash    []byte
}

func NewStateCache(stateRoot string) *StateCache {
	return &StateCache{
		stateRoot: stateRoot,
		txs:       make([]*tx.Tx, 0),

		pendingNewAccountIndexMap:      make(map[int64]int, 0),
		pendingNewLiquidityIndexMap:    make(map[int64]int, 0),
		pendingNewNftIndexMap:          make(map[int64]int, 0),
		pendingUpdateAccountIndexMap:   make(map[int64]int, 0),
		pendingUpdateLiquidityIndexMap: make(map[int64]int, 0),
		pendingUpdateNftIndexMap:       make(map[int64]int, 0),

		pubData:                         make([]byte, 0),
		priorityOperations:              0,
		pubDataOffset:                   make([]uint32, 0),
		pendingOnChainOperationsPubData: make([][]byte, 0),
		pendingOnChainOperationsHash:    common.FromHex(util.EmptyStringKeccak),
	}
}

type StatesToCommit struct {
	Block                        *block.Block
	BlockForCommit               *blockForCommit.BlockForCommit
	PendingNewAccount            []*account.Account
	PendingUpdateAccount         []*account.Account
	PendingNewAccountHistory     []*account.AccountHistory
	PendingNewLiquidity          []*liquidity.Liquidity
	PendingUpdateLiquidity       []*liquidity.Liquidity
	PendingNewLiquidityHistory   []*liquidity.LiquidityHistory
	PendingNewNft                []*nft.L2Nft
	PendingUpdateNft             []*nft.L2Nft
	PendingNewNftHistory         []*nft.L2NftHistory
	PendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory
}

type BlockChain struct {
	BlockModel            block.BlockModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	L2AssetInfoModel      asset.AssetModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	L2NftHistoryModel     nft.L2NftHistoryModel

	accountMap        map[int64]*commonAsset.AccountInfo
	liquidityMap      map[int64]*liquidity.Liquidity
	nftMap            map[int64]*nft.L2Nft
	accountTree       bsmt.SparseMerkleTree
	liquidityTree     bsmt.SparseMerkleTree
	nftTree           bsmt.SparseMerkleTree
	accountAssetTrees []bsmt.SparseMerkleTree

	chainConfig *ChainConfig
	redisCache  dbcache.Cache
	processor   Processor

	currentBlock *block.Block
	stateCache   *StateCache // Cache for current block changes.
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	gormPointer, err := gorm.Open(postgres.Open(config.Postgres.DataSource))
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	conn := sqlx.NewSqlConn("postgres", config.Postgres.DataSource)

	bc := &BlockChain{
		accountMap:   make(map[int64]*commonAsset.AccountInfo),
		liquidityMap: make(map[int64]*liquidity.Liquidity),
		nftMap:       make(map[int64]*nft.L2Nft),

		BlockModel:            block.NewBlockModel(conn, config.CacheRedis, gormPointer),
		TxModel:               tx.NewTxModel(conn, config.CacheRedis, gormPointer),
		TxDetailModel:         tx.NewTxDetailModel(conn, config.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, config.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, config.CacheRedis, gormPointer),
		L2AssetInfoModel:      asset.NewAssetModel(conn, config.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, config.CacheRedis, gormPointer),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, config.CacheRedis, gormPointer),
		L2NftModel:            nft.NewL2NftModel(conn, config.CacheRedis, gormPointer),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(conn, config.CacheRedis, gormPointer),

		chainConfig: config,
		redisCache:  dbcache.NewRedisCache(config.CacheRedis[0].Host, config.CacheRedis[0].Pass, 15*time.Minute),
	}

	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block failed: ", err)
		return nil, err
	}
	bc.currentBlock, err = bc.BlockModel.GetBlockByBlockHeight(curHeight)
	if err != nil {
		return nil, err
	}
	bc.stateCache = NewStateCache(bc.currentBlock.StateRoot)
	if bc.currentBlock.BlockStatus == block.StatusProposing {
		curHeight--
	}

	treeCtx := &treedb.Context{
		Name:          moduleName,
		Driver:        config.TreeDB.Driver,
		LevelDBOption: &config.TreeDB.LevelDBOption,
		RedisDBOption: &config.TreeDB.RedisDBOption,
	}
	err = treedb.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Error("setup tree db failed: ", err)
		return nil, err
	}
	bc.accountTree, bc.accountAssetTrees, err = tree.InitAccountTree(
		bc.AccountModel,
		bc.AccountHistoryModel,
		curHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("init account tree failed:", err)
		return nil, err
	}
	bc.liquidityTree, err = tree.InitLiquidityTree(
		bc.LiquidityHistoryModel,
		curHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("init liquidity tree failed:", err)
		return nil, err
	}
	bc.nftTree, err = tree.InitNftTree(
		bc.L2NftHistoryModel,
		curHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("init nft tree failed:", err)
		return nil, err
	}

	bc.processor = NewCommitProcessor(bc)
	return bc, nil
}

func (bc *BlockChain) GetPendingTxs() []*tx.Tx {
	return bc.stateCache.txs
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx) (*tx.Tx, error) {
	return bc.processor.Process(tx)
}

func (bc *BlockChain) SyncStateCacheToRedis() error {
	stateCache := bc.stateCache

	// Sync new create to cache.
	for accountIndex, status := range stateCache.pendingNewAccountIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(accountIndex), bc.accountMap[accountIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingNewAccountIndexMap[accountIndex] = StateCacheCached
	}
	for liquidityIndex, status := range stateCache.pendingNewLiquidityIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.LiquidityKeyByIndex(liquidityIndex), bc.liquidityMap[liquidityIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingNewLiquidityIndexMap[liquidityIndex] = StateCacheCached
	}
	for nftIndex, status := range stateCache.pendingNewNftIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.NftKeyByIndex(nftIndex), bc.nftMap[nftIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingNewNftIndexMap[nftIndex] = StateCacheCached
	}

	// Sync pending update to cache.
	for accountIndex, status := range stateCache.pendingUpdateAccountIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(accountIndex), bc.accountMap[accountIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingUpdateAccountIndexMap[accountIndex] = StateCacheCached
	}
	for liquidityIndex, status := range stateCache.pendingUpdateLiquidityIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.LiquidityKeyByIndex(liquidityIndex), bc.liquidityMap[liquidityIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingUpdateLiquidityIndexMap[liquidityIndex] = StateCacheCached
	}
	for nftIndex, status := range stateCache.pendingUpdateNftIndexMap {
		if status != StateCachePending {
			continue
		}

		err := bc.redisCache.Set(context.Background(), dbcache.NftKeyByIndex(nftIndex), bc.nftMap[nftIndex])
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		stateCache.pendingUpdateNftIndexMap[nftIndex] = StateCacheCached
	}

	return nil
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
	bc.stateCache = NewStateCache(bc.currentBlock.StateRoot)
	return newBlock, nil
}

func (bc *BlockChain) CommitNewBlock(blockSize int, createdAt int64) (*StatesToCommit, error) {
	newBlock, newBlockForCommit, err := bc.commitNewBlock(blockSize, createdAt)
	if err != nil {
		return nil, err
	}

	pendingNewAccount, pendingUpdateAccount, pendingNewAccountHistory, err := bc.getPendingAccount()
	if err != nil {
		return nil, err
	}

	pendingNewLiquidity, pendingUpdateLiquidity, pendingNewLiquidityHistory, err := bc.getPendingLiquidity()
	if err != nil {
		return nil, err
	}

	pendingNewNft, pendingUpdateNft, pendingNewNftHistory, err := bc.getPendingNft()
	if err != nil {
		return nil, err
	}

	return &StatesToCommit{
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
		PendingNewNftWithdrawHistory: bc.stateCache.pendingNewNftWithdrawHistory,
	}, nil
}

func (bc *BlockChain) commitNewBlock(blockSize int, createdAt int64) (*block.Block, *blockForCommit.BlockForCommit, error) {
	stateCache := bc.stateCache
	if blockSize < len(stateCache.txs) {
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

	commitment := util.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(newBlock.StateRoot), common.FromHex(stateCache.stateRoot),
		stateCache.pubData, int64(len(stateCache.pubDataOffset)))
	newBlock.BlockSize = uint16(blockSize)
	newBlock.BlockCommitment = commitment
	newBlock.StateRoot = stateCache.stateRoot
	newBlock.PriorityOperations = stateCache.priorityOperations
	newBlock.PendingOnChainOperationsHash = common.Bytes2Hex(stateCache.pendingOnChainOperationsHash)
	newBlock.Txs = stateCache.txs
	newBlock.BlockStatus = block.StatusPending
	if stateCache.pendingOnChainOperationsPubData != nil {
		onChainOperationsPubDataBytes, err := json.Marshal(stateCache.pendingOnChainOperationsPubData)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(stateCache.pubDataOffset)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}
	newBlockForCommit := &blockForCommit.BlockForCommit{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(stateCache.pubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
	}

	bc.currentBlock = newBlock
	return newBlock, newBlockForCommit, nil
}

func (bc *BlockChain) getPendingAccount() ([]*account.Account, []*account.Account, []*account.AccountHistory, error) {
	stateCache := bc.stateCache
	pendingNewAccount := make([]*account.Account, 0)
	pendingUpdateAccount := make([]*account.Account, 0)
	pendingNewAccountHistory := make([]*account.AccountHistory, 0)

	for index, status := range stateCache.pendingNewAccountIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newAccount, err := commonAsset.FromFormatAccountInfo(bc.accountMap[index])
		if err != nil {
			return nil, nil, nil, err
		}

		pendingNewAccount = append(pendingNewAccount, newAccount)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   bc.currentBlock.BlockHeight, // TODO: ensure this should be the new block's height.
		})
	}

	for index, status := range stateCache.pendingUpdateAccountIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newAccount, err := commonAsset.FromFormatAccountInfo(bc.accountMap[index])
		if err != nil {
			return nil, nil, nil, err
		}
		pendingUpdateAccount = append(pendingUpdateAccount, newAccount)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   bc.currentBlock.BlockHeight, // TODO: ensure this should be the new block's height.
		})
	}

	return pendingNewAccount, pendingUpdateAccount, pendingNewAccountHistory, nil
}

func (bc *BlockChain) getPendingLiquidity() ([]*liquidity.Liquidity, []*liquidity.Liquidity, []*liquidity.LiquidityHistory, error) {
	stateCache := bc.stateCache
	pendingNewLiquidity := make([]*liquidity.Liquidity, 0)
	pendingUpdateLiquidity := make([]*liquidity.Liquidity, 0)
	pendingNewLiquidityHistory := make([]*liquidity.LiquidityHistory, 0)

	for index, status := range stateCache.pendingNewLiquidityIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newLiquidity := bc.liquidityMap[index]
		pendingNewLiquidity = append(pendingNewLiquidity, newLiquidity)
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            newLiquidity.PairIndex,
			AssetAId:             newLiquidity.AssetAId,
			AssetA:               newLiquidity.AssetA,
			AssetBId:             newLiquidity.AssetBId,
			AssetB:               newLiquidity.AssetB,
			LpAmount:             newLiquidity.LpAmount,
			KLast:                newLiquidity.KLast,
			FeeRate:              newLiquidity.FeeRate,
			TreasuryAccountIndex: newLiquidity.TreasuryAccountIndex,
			TreasuryRate:         newLiquidity.TreasuryRate,
			L2BlockHeight:        bc.currentBlock.BlockHeight,
		})
	}

	for index, status := range stateCache.pendingUpdateLiquidityIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newLiquidity := bc.liquidityMap[index]
		pendingUpdateLiquidity = append(pendingUpdateLiquidity, newLiquidity)
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            newLiquidity.PairIndex,
			AssetAId:             newLiquidity.AssetAId,
			AssetA:               newLiquidity.AssetA,
			AssetBId:             newLiquidity.AssetBId,
			AssetB:               newLiquidity.AssetB,
			LpAmount:             newLiquidity.LpAmount,
			KLast:                newLiquidity.KLast,
			FeeRate:              newLiquidity.FeeRate,
			TreasuryAccountIndex: newLiquidity.TreasuryAccountIndex,
			TreasuryRate:         newLiquidity.TreasuryRate,
			L2BlockHeight:        bc.currentBlock.BlockHeight,
		})
	}

	return pendingNewLiquidity, pendingUpdateLiquidity, pendingNewLiquidityHistory, nil
}

func (bc *BlockChain) getPendingNft() ([]*nft.L2Nft, []*nft.L2Nft, []*nft.L2NftHistory, error) {
	stateCache := bc.stateCache
	pendingNewNft := make([]*nft.L2Nft, 0)
	pendingUpdateNft := make([]*nft.L2Nft, 0)
	pendingNewNftHistory := make([]*nft.L2NftHistory, 0)

	for index, status := range stateCache.pendingNewNftIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newNft := bc.nftMap[index]
		pendingNewNft = append(pendingNewNft, newNft)
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
			NftIndex:            newNft.NftIndex,
			CreatorAccountIndex: newNft.CreatorAccountIndex,
			OwnerAccountIndex:   newNft.OwnerAccountIndex,
			NftContentHash:      newNft.NftContentHash,
			NftL1Address:        newNft.NftL1Address,
			NftL1TokenId:        newNft.NftL1TokenId,
			CreatorTreasuryRate: newNft.CreatorTreasuryRate,
			CollectionId:        newNft.CollectionId,
			L2BlockHeight:       bc.currentBlock.BlockHeight,
		})
	}

	for index, status := range stateCache.pendingUpdateNftIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in state cache")
			continue
		}

		newNft := bc.nftMap[index]
		pendingUpdateNft = append(pendingUpdateNft, newNft)
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
			NftIndex:            newNft.NftIndex,
			CreatorAccountIndex: newNft.CreatorAccountIndex,
			OwnerAccountIndex:   newNft.OwnerAccountIndex,
			NftContentHash:      newNft.NftContentHash,
			NftL1Address:        newNft.NftL1Address,
			NftL1TokenId:        newNft.NftL1TokenId,
			CreatorTreasuryRate: newNft.CreatorTreasuryRate,
			CollectionId:        newNft.CollectionId,
			L2BlockHeight:       bc.currentBlock.BlockHeight,
		})
	}

	return pendingNewNft, pendingUpdateNft, pendingNewNftHistory, nil
}

func (bc *BlockChain) getNextAccountIndex() int64 {
	return int64(len(bc.accountAssetTrees)) + 1
}

func (bc *BlockChain) getNextNftIndex() (int64, error) {
	stateCache := bc.stateCache
	if stateCache == nil || len(stateCache.pendingNewNftIndexMap) == 0 {
		maxNftIndex, err := bc.L2NftModel.GetLatestNftIndex()
		if err != nil {
			return -1, err
		}
		return maxNftIndex + 1, nil
	}

	maxNftIndex := int64(-1)
	for index, status := range stateCache.pendingNewNftIndexMap {
		if status >= StateCachePending && index > maxNftIndex {
			maxNftIndex = index
		}
	}
	return maxNftIndex + 1, nil
}

func (bc *BlockChain) prepareAccountsAndAssets(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		if bc.accountMap[accountIndex] == nil {
			accountInfo, err := bc.AccountModel.GetAccountByAccountIndex(accountIndex)
			if err != nil {
				return err
			}
			bc.accountMap[accountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
			if err != nil {
				return fmt.Errorf("convert to format account info failed: %v", err)
			}
		}
		if bc.accountMap[accountIndex].AssetInfo == nil {
			bc.accountMap[accountIndex].AssetInfo = make(map[int64]*commonAsset.AccountAsset)
		}
		for _, assetId := range assets {
			if bc.accountMap[accountIndex].AssetInfo[assetId] == nil {
				bc.accountMap[accountIndex].AssetInfo[assetId] = &commonAsset.AccountAsset{
					AssetId:                  assetId,
					Balance:                  ZeroBigInt,
					LpAmount:                 ZeroBigInt,
					OfferCanceledOrFinalized: ZeroBigInt,
				}
			}
		}
	}

	return nil
}

func (bc *BlockChain) prepareLiquidity(pairIndex int64) error {
	if bc.liquidityMap[pairIndex] == nil {
		liquidityInfo, err := bc.LiquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return err
		}
		bc.liquidityMap[pairIndex] = liquidityInfo
	}
	return nil
}

func (bc *BlockChain) prepareNft(nftIndex int64) error {
	if bc.nftMap[nftIndex] == nil {
		nftAsset, err := bc.L2NftModel.GetNftAsset(nftIndex)
		if err != nil {
			return err
		}
		bc.nftMap[nftIndex] = nftAsset
	}
	return nil
}

func (bc *BlockChain) updateAccountTree(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		for _, assetId := range assets {
			assetLeaf, err := tree.ComputeAccountAssetLeafHash(
				bc.accountMap[accountIndex].AssetInfo[assetId].Balance.String(),
				bc.accountMap[accountIndex].AssetInfo[assetId].LpAmount.String(),
				bc.accountMap[accountIndex].AssetInfo[assetId].OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				return fmt.Errorf("compute new account asset leaf failed: %v", err)
			}
			err = bc.accountAssetTrees[accountIndex].Set(uint64(assetId), assetLeaf)
			if err != nil {
				return fmt.Errorf("update asset tree failed: %v", err)
			}
		}

		bc.accountMap[accountIndex].AssetRoot = common.Bytes2Hex(bc.accountAssetTrees[accountIndex].Root())
		nAccountLeafHash, err := tree.ComputeAccountLeafHash(
			bc.accountMap[accountIndex].AccountNameHash,
			bc.accountMap[accountIndex].PublicKey,
			bc.accountMap[accountIndex].Nonce,
			bc.accountMap[accountIndex].CollectionNonce,
			bc.accountAssetTrees[accountIndex].Root(),
		)
		if err != nil {
			return fmt.Errorf("unable to compute account leaf: %v", err)
		}
		err = bc.accountTree.Set(uint64(accountIndex), nAccountLeafHash)
		if err != nil {
			return fmt.Errorf("unable to update account tree: %v", err)
		}
	}

	return nil
}

func (bc *BlockChain) updateLiquidityTree(pairIndex int64) error {
	nLiquidityAssetLeaf, err := tree.ComputeLiquidityAssetLeafHash(
		bc.liquidityMap[pairIndex].AssetAId,
		bc.liquidityMap[pairIndex].AssetA,
		bc.liquidityMap[pairIndex].AssetBId,
		bc.liquidityMap[pairIndex].AssetB,
		bc.liquidityMap[pairIndex].LpAmount,
		bc.liquidityMap[pairIndex].KLast,
		bc.liquidityMap[pairIndex].FeeRate,
		bc.liquidityMap[pairIndex].TreasuryAccountIndex,
		bc.liquidityMap[pairIndex].TreasuryRate,
	)
	if err != nil {
		return fmt.Errorf("unable to compute liquidity leaf: %v", err)
	}
	err = bc.liquidityTree.Set(uint64(pairIndex), nLiquidityAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update liquidity tree: %v", err)
	}

	return nil
}

func (bc *BlockChain) updateNftTree(nftIndex int64) error {
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		bc.nftMap[nftIndex].CreatorAccountIndex,
		bc.nftMap[nftIndex].OwnerAccountIndex,
		bc.nftMap[nftIndex].NftContentHash,
		bc.nftMap[nftIndex].NftL1Address,
		bc.nftMap[nftIndex].NftL1TokenId,
		bc.nftMap[nftIndex].CreatorTreasuryRate,
		bc.nftMap[nftIndex].CollectionId,
	)
	if err != nil {
		return fmt.Errorf("unable to compute nft leaf: %v", err)
	}
	err = bc.nftTree.Set(uint64(nftIndex), nftAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update nft tree: %v", err)
	}

	return nil
}

func (bc *BlockChain) getStateRoot() string {
	hFunc := mimc.NewMiMC()
	hFunc.Write(bc.accountTree.Root())
	hFunc.Write(bc.liquidityTree.Root())
	hFunc.Write(bc.nftTree.Root())
	return common.Bytes2Hex(hFunc.Sum(nil))
}
