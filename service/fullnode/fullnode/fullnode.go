package fullnode

import (
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb/dao/block"
	"gorm.io/gorm"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-go-sdk/client"
	"github.com/bnb-chain/zkbnb/core"
	tx "github.com/bnb-chain/zkbnb/dao/tx"
)

const (
	MaxFullnodeInterval = 60 * 1
	DefaultL2EndPoint   = "http://localhost:8888"
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes []int
	}

	ApiServer struct {
		L2EndPoint string
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
	// TODO: add BlockSize in zkbnb-go-sdk

	curHeight, err := c.bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic("get current block height failed: " + err.Error())
	}

	curBlock, err := c.bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		panic("get current block failed: " + err.Error())
	}

	for {
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock, err = c.bc.ProposeNewBlock()
			if err != nil {
				panic("propose new block failed: " + err.Error())
			}

			curHeight++
		}

		l2Block, err := c.client.GetBlockByHeight(curHeight)
		if err != nil {
			// TODO check error
			logx.Errorf("get block failed, height: %d, err %v ", curHeight, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		txs := make([]*tx.Tx, 0, len(l2Block.Txs))

		for _, blockTx := range l2Block.Txs {
			newTx := &tx.Tx{
				TxHash: blockTx.Hash, // Would be computed in prepare method of executors.
				TxType: blockTx.Type,
				TxInfo: blockTx.Info,
			}

			txs = append(txs, newTx)
			err = c.bc.ApplyTransaction(newTx)
			if err != nil {
				logx.Errorf("apply block tx ID: %d failed, err %v ", newTx.ID, err)
				continue
			}
		}

		if c.bc.Statedb.StateRoot != l2Block.StateRoot {
			panic(fmt.Sprintf("state root not matched between statedb and l2block: %d", l2Block.Height))
		}

		logx.Infof("commit new block on fullnode, height=%d, blockSize=%d", curBlock.BlockHeight, curBlock.BlockSize)
		curBlock, err = c.commitNewBlock(curBlock)
		if err != nil {
			panic("commit new block failed: " + err.Error())
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Fullnode) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	blockSize := curBlock.BlockSize
	blockStates, err := c.bc.CommitNewBlock(int(blockSize), curBlock.CreatedAt.UnixMilli())
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
