package core

import (
	"fmt"
	"math/big"
	"time"

	"github.com/bnb-chain/zkbas/common/util"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/pkg/treedb"
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

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

const (
	_ = iota
	StateCachePending
	StateCacheCached
)

type StateCache struct {
	blockNumber int64
	txs         []*tx.Tx

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

func NewStateCache(blockNumber int64) *StateCache {
	return &StateCache{
		blockNumber: blockNumber,
		txs:         make([]*tx.Tx, 0),

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

func (s *StateCache) GetTxs() []*tx.Tx {
	return s.txs
}

type BlockChain struct {
	AccountMap   map[int64]*commonAsset.AccountInfo
	LiquidityMap map[int64]*liquidity.Liquidity
	NftMap       map[int64]*nft.L2Nft

	accountTree       bsmt.SparseMerkleTree
	liquidityTree     bsmt.SparseMerkleTree
	nftTree           bsmt.SparseMerkleTree
	accountAssetTrees []bsmt.SparseMerkleTree

	BlockModel            block.BlockModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	L2AssetInfoModel      assetInfo.AssetInfoModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	L2NftHistoryModel     nft.L2NftHistoryModel

	currentBlock int64

	chainConfig *ChainConfig
	processor   Processor
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	gormPointer, err := gorm.Open(postgres.Open(config.Postgres.DataSource))
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	conn := sqlx.NewSqlConn("postgres", config.Postgres.DataSource)
	redisConn := redis.New(config.CacheRedis[0].Host, WithRedis(config.CacheRedis[0].Type, config.CacheRedis[0].Pass))

	bc := &BlockChain{
		AccountMap:   make(map[int64]*commonAsset.AccountInfo),
		LiquidityMap: make(map[int64]*liquidity.Liquidity),
		NftMap:       make(map[int64]*nft.L2Nft),

		BlockModel:            block.NewBlockModel(conn, config.CacheRedis, gormPointer, redisConn),
		TxModel:               tx.NewTxModel(conn, config.CacheRedis, gormPointer, redisConn),
		TxDetailModel:         tx.NewTxDetailModel(conn, config.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, config.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, config.CacheRedis, gormPointer),
		L2AssetInfoModel:      assetInfo.NewAssetInfoModel(conn, config.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, config.CacheRedis, gormPointer),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, config.CacheRedis, gormPointer),
		L2NftModel:            nft.NewL2NftModel(conn, config.CacheRedis, gormPointer),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(conn, config.CacheRedis, gormPointer),

		chainConfig: config,
	}

	bc.currentBlock, err = bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Error("get current block failed: ", err)
		return nil, err
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
		bc.currentBlock,
		treeCtx,
	)
	if err != nil {
		logx.Error("init account tree failed:", err)
		return nil, err
	}
	bc.liquidityTree, err = tree.InitLiquidityTree(
		bc.LiquidityHistoryModel,
		bc.currentBlock,
		treeCtx,
	)
	if err != nil {
		logx.Error("init liquidity tree failed:", err)
		return nil, err
	}
	bc.nftTree, err = tree.InitNftTree(
		bc.L2NftHistoryModel,
		bc.currentBlock,
		treeCtx,
	)
	if err != nil {
		logx.Error("init nft tree failed:", err)
		return nil, err
	}

	bc.processor = NewCommitProcessor(bc)
	return bc, nil
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx, stateCache *StateCache) (*tx.Tx, *StateCache, error) {
	return bc.processor.Process(tx, stateCache)
}

func (bc *BlockChain) SyncCache(stateCache *StateCache) error {
	// Iterate pending state, sync them to the cache.
	return nil
}

func (bc *BlockChain) ProposeNewBlock() (*block.Block, error) {
	createdAt := time.Now().UnixMilli()
	newBlock := &block.Block{
		Model: gorm.Model{
			CreatedAt: time.UnixMilli(createdAt),
		},
		BlockHeight: bc.currentBlock + 1,
		BlockStatus: block.StatusProposing,
	}

	err := bc.BlockModel.CreateNewBlock(newBlock)
	if err != nil {
		return nil, err
	}

	bc.currentBlock++
	return newBlock, nil
}

func (bc *BlockChain) CommitNewBlock(stateCache *StateCache) error {
	return nil
}

func (bc *BlockChain) prepareAccountsAndAssets(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		if bc.AccountMap[accountIndex] == nil {
			accountInfo, err := bc.AccountModel.GetAccountByAccountIndex(accountIndex)
			if err != nil {
				return err
			}
			bc.AccountMap[accountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
			if err != nil {
				return fmt.Errorf("convert to format account info failed: %v", err)
			}
		}
		if bc.AccountMap[accountIndex].AssetInfo == nil {
			bc.AccountMap[accountIndex].AssetInfo = make(map[int64]*commonAsset.AccountAsset)
		}
		for _, assetId := range assets {
			if bc.AccountMap[accountIndex].AssetInfo[assetId] == nil {
				bc.AccountMap[accountIndex].AssetInfo[assetId] = &commonAsset.AccountAsset{
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
	if bc.LiquidityMap[pairIndex] == nil {
		liquidityInfo, err := bc.LiquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return err
		}
		bc.LiquidityMap[pairIndex] = liquidityInfo
	}
	return nil
}

func (bc *BlockChain) prepareNft(nftIndex int64) error {
	if bc.NftMap[nftIndex] == nil {
		nftAsset, err := bc.L2NftModel.GetNftAsset(nftIndex)
		if err != nil {
			return err
		}
		bc.NftMap[nftIndex] = nftAsset
	}
	return nil
}

func (bc *BlockChain) updateAccountTree(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		for _, assetId := range assets {
			assetLeaf, err := tree.ComputeAccountAssetLeafHash(
				bc.AccountMap[accountIndex].AssetInfo[assetId].Balance.String(),
				bc.AccountMap[accountIndex].AssetInfo[assetId].LpAmount.String(),
				bc.AccountMap[accountIndex].AssetInfo[assetId].OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				return fmt.Errorf("compute new account asset leaf failed: %v", err)
			}
			err = bc.accountAssetTrees[accountIndex].Set(uint64(assetId), assetLeaf)
			if err != nil {
				return fmt.Errorf("update asset tree failed: %v", err)
			}
		}

		bc.AccountMap[accountIndex].AssetRoot = common.Bytes2Hex(bc.accountAssetTrees[accountIndex].Root())
		nAccountLeafHash, err := tree.ComputeAccountLeafHash(
			bc.AccountMap[accountIndex].AccountNameHash,
			bc.AccountMap[accountIndex].PublicKey,
			bc.AccountMap[accountIndex].Nonce,
			bc.AccountMap[accountIndex].CollectionNonce,
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
		bc.LiquidityMap[pairIndex].AssetAId,
		bc.LiquidityMap[pairIndex].AssetA,
		bc.LiquidityMap[pairIndex].AssetBId,
		bc.LiquidityMap[pairIndex].AssetB,
		bc.LiquidityMap[pairIndex].LpAmount,
		bc.LiquidityMap[pairIndex].KLast,
		bc.LiquidityMap[pairIndex].FeeRate,
		bc.LiquidityMap[pairIndex].TreasuryAccountIndex,
		bc.LiquidityMap[pairIndex].TreasuryRate,
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
		bc.NftMap[nftIndex].CreatorAccountIndex,
		bc.NftMap[nftIndex].OwnerAccountIndex,
		bc.NftMap[nftIndex].NftContentHash,
		bc.NftMap[nftIndex].NftL1Address,
		bc.NftMap[nftIndex].NftL1TokenId,
		bc.NftMap[nftIndex].CreatorTreasuryRate,
		bc.NftMap[nftIndex].CollectionId,
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
