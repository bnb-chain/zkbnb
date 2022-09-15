package fullnode

import (
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-go-sdk/client"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/tx"
)

const (
	MaxFullnodeInterval = 60 * 1
	DefaultL2EndPoint          = "http://localhost:8888"
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes []int
	}

	ApiServer  struct {
		L2EndPoint string `json:",optional"`
	}
}

type Fullnode struct {
	config *Config
	client client.ZkBNBClient
	bc     *core.BlockChain
}

func NewFullnode(config *Config) (*Fullnode, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, errors.New("nil optional block sizes")
	}

	bc, err := core.NewBlockChain(&config.ChainConfig, "fullnode")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	l2EndPoint := config.ApiServer.L2EndPoint
	if len(l2EndPoint) == 0 {
		l2EndPoint = DefaultL2EndPoint
	}

	fullnode := &Fullnode{
		config: config,
		client: client.NewZkBNBClient(l2EndPoint),
		bc:     bc,
	}
	return fullnode, nil
}

func (c *Fullnode) Run() {
	curHeight, err := c.bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic("get current block height failed: " + err.Error())
	}
	curHeight++

	for {
		l2Block, err := c.client.GetBlockByHeight(curHeight)
		if err != nil {
			logx.Errorf("get block failed, height: %d, err %v ", curHeight, err)
			continue
		}

		for _, blockTx := range l2Block.Txs {
			newTx := &tx.Tx{
				TxHash: blockTx.Hash, // Would be computed in prepare method of executors.
				TxType: blockTx.Type,
				TxInfo: blockTx.Info,

				GasFeeAssetId: blockTx.GasFeeAssetId,
				GasFee:        blockTx.GasFee,
				PairIndex:     blockTx.PairIndex,
				NftIndex:      blockTx.NftIndex,
				CollectionId:  blockTx.CollectionId,
				AssetId:       blockTx.AssetId,
				TxAmount:      blockTx.Amount,
				NativeAddress: blockTx.NativeAddress,

				BlockHeight: blockTx.BlockHeight,
				TxStatus:    int(blockTx.Status),
			}

			err = c.bc.ApplyTransaction(newTx)
			if err != nil {
				logx.Errorf("apply block tx ID: %d failed, err %v ", newTx.ID, err)
				continue
			}
		}

		err = c.bc.StateDB().SyncStateCacheToRedis()
		if err != nil {
			panic("sync redis cache failed: " + err.Error())
		}

		curHeight++
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Fullnode) restoreExecutedTxs() (*block.Block, error) {
	bc := c.bc
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, err
	}
	curBlock, err := bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}

	executedTxs, err := c.bc.TxPoolModel.GetTxsByStatus(tx.StatusExecuted)
	if err != nil {
		return nil, err
	}

	if curBlock.BlockStatus > block.StatusProposing {
		if len(executedTxs) != 0 {
			return nil, errors.New("no proposing block but exist executed txs")
		}
		return curBlock, nil
	}

	for _, executedTx := range executedTxs {
		err = c.bc.ApplyTransaction(executedTx)
		if err != nil {
			return nil, err
		}
	}

	return curBlock, nil
}

func (c *Fullnode) createNewBlock(curBlock *block.Block, poolTx *tx.Tx) error {
	return c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		err := c.bc.TxPoolModel.UpdateTxsInTransact(dbTx, []*tx.Tx{poolTx})
		if err != nil {
			return err
		}

		return c.bc.BlockModel.CreateBlockInTransact(dbTx, curBlock)
	})
}

func (c *Fullnode) shouldCommit(curBlock *block.Block) bool {
	var now = time.Now()
	if (len(c.bc.Statedb.Txs) > 0 && now.Unix()-curBlock.CreatedAt.Unix() >= MaxFullnodeInterval) ||
		len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
		return true
	}

	return false
}

func (c *Fullnode) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	blockSize := c.computeCurrentBlockSize()
	blockStates, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
	if err != nil {
		return nil, err
	}

	// update db
	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		// create block for commit
		if blockStates.CompressedBlock != nil {
			err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			if err != nil {
				return err
			}
		}
		// create new account
		if len(blockStates.PendingNewAccount) != 0 {
			err = c.bc.DB().AccountModel.CreateAccountsInTransact(tx, blockStates.PendingNewAccount)
			if err != nil {
				return err
			}
		}
		// update account
		if len(blockStates.PendingUpdateAccount) != 0 {
			err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingUpdateAccount)
			if err != nil {
				return err
			}
		}
		// create new account history
		if len(blockStates.PendingNewAccountHistory) != 0 {
			err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingNewAccountHistory)
			if err != nil {
				return err
			}
		}
		// create new liquidity
		if len(blockStates.PendingNewLiquidity) != 0 {
			err = c.bc.DB().LiquidityModel.CreateLiquidityInTransact(tx, blockStates.PendingNewLiquidity)
			if err != nil {
				return err
			}
		}
		// update liquidity
		if len(blockStates.PendingUpdateLiquidity) != 0 {
			err = c.bc.DB().LiquidityModel.UpdateLiquidityInTransact(tx, blockStates.PendingUpdateLiquidity)
			if err != nil {
				return err
			}
		}
		// create new liquidity history
		if len(blockStates.PendingNewLiquidityHistory) != 0 {
			err = c.bc.DB().LiquidityHistoryModel.CreateLiquidityHistoriesInTransact(tx, blockStates.PendingNewLiquidityHistory)
			if err != nil {
				return err
			}
		}
		// create new nft
		if len(blockStates.PendingNewNft) != 0 {
			err = c.bc.DB().L2NftModel.CreateNftsInTransact(tx, blockStates.PendingNewNft)
			if err != nil {
				return err
			}
		}
		// update nft
		if len(blockStates.PendingUpdateNft) != 0 {
			err = c.bc.DB().L2NftModel.UpdateNftsInTransact(tx, blockStates.PendingUpdateNft)
			if err != nil {
				return err
			}
		}
		// new nft history
		if len(blockStates.PendingNewNftHistory) != 0 {
			err = c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(tx, blockStates.PendingNewNftHistory)
			if err != nil {
				return err
			}
		}
		// delete txs from tx pool
		err := c.bc.DB().TxPoolModel.DeleteTxsInTransact(tx, blockStates.Block.Txs)
		if err != nil {
			return err
		}
		// update block
		blockStates.Block.ClearTxsModel()
		return c.bc.DB().BlockModel.UpdateBlockInTransact(tx, blockStates.Block)
	})

	if err != nil {
		return nil, err
	}

	return blockStates.Block, nil
}

func (c *Fullnode) computeCurrentBlockSize() int {
	var blockSize int
	for i := 0; i < len(c.optionalBlockSizes); i++ {
		if len(c.bc.Statedb.Txs) <= c.optionalBlockSizes[i] {
			blockSize = c.optionalBlockSizes[i]
			break
		}
	}
	return blockSize
}
