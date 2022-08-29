package committer

import (
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/core"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
)

const (
	MaxCommitterInterval = 1
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes []int
	}
}

type Committer struct {
	config             *Config
	maxTxsPerBlock     int
	optionalBlockSizes []int

	bc *core.BlockChain

	memPoolModel       mempool.MempoolModel
	executedMemPoolTxs []*mempool.MempoolTx
}

func NewCommitter(config *Config) (*Committer, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, errors.New("nil optional block sizes")
	}

	bc, err := core.NewBlockChain(&config.ChainConfig, "committer")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	gormPointer, err := gorm.Open(postgres.Open(config.Postgres.DataSource))
	if err != nil {
		logx.Error("gorm connect db failed: ", err)
		return nil, err
	}
	conn := sqlx.NewSqlConn("postgres", config.Postgres.DataSource)

	committer := &Committer{
		config:             config,
		maxTxsPerBlock:     config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		optionalBlockSizes: config.BlockConfig.OptionalBlockSizes,

		bc: bc,

		memPoolModel:       mempool.NewMempoolModel(conn, config.CacheRedis, gormPointer),
		executedMemPoolTxs: make([]*mempool.MempoolTx, 0),
	}
	return committer, nil
}

func (c *Committer) Run() {
	curBlock, err := c.restoreExecutedTxs()
	if err != nil {
		panic("restore executed tx failed: " + err.Error())
	}

	for {
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock, err = c.bc.ProposeNewBlock()
			if err != nil {
				panic("propose new block failed: " + err.Error())
			}
		}

		// Read pending transactions from mempool_tx table.
		pendingTxs, err := c.memPoolModel.GetMempoolTxsByStatus(mempool.PendingTxStatus)
		if err != nil {
			logx.Error("get pending transactions from mempool failed:", err)
			return
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs, err = c.memPoolModel.GetMempoolTxsByStatus(mempool.PendingTxStatus)
			if err != nil {
				logx.Error("get pending transactions from mempool failed:", err)
				return
			}
		}

		pendingUpdateMempoolTxs := make([]*mempool.MempoolTx, 0, len(pendingTxs))
		pendingDeleteMempoolTxs := make([]*mempool.MempoolTx, 0, len(pendingTxs))
		for _, mempoolTx := range pendingTxs {
			if c.shouldCommit(curBlock) {
				break
			}

			tx := convertMempoolTxToTx(mempoolTx)
			err = c.bc.ApplyTransaction(tx)
			if err != nil {
				logx.Errorf("Apply mempool tx (ID: ", mempoolTx.ID, ") failed: ", err)
				mempoolTx.Status = mempool.FailTxStatus
				pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
				continue
			}
			mempoolTx.Status = mempool.ExecutedTxStatus
			pendingUpdateMempoolTxs = append(pendingUpdateMempoolTxs, mempoolTx)

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				err = c.createNewBlock(curBlock)
				if err != nil {
					panic("create new block failed" + err.Error())
				}
			}
		}

		err = c.bc.StateDB().SyncStateCacheToRedis()
		if err != nil {
			panic("sync redis cache failed: " + err.Error())
		}

		err = c.memPoolModel.UpdateMempoolTxs(pendingUpdateMempoolTxs, pendingDeleteMempoolTxs)
		if err != nil {
			panic("update mempool failed: " + err.Error())
		}
		c.executedMemPoolTxs = append(c.executedMemPoolTxs, pendingUpdateMempoolTxs...)

		if c.shouldCommit(curBlock) {
			curBlock, err = c.commitNewBlock(curBlock)
			if err != nil {
				panic("commit new block failed: " + err.Error())
			}
		}
	}
}

func (c *Committer) restoreExecutedTxs() (*block.Block, error) {
	bc := c.bc
	curHeight, err := bc.BlockModel.GetCurrentHeight()
	if err != nil {
		return nil, err
	}
	curBlock, err := bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}

	executedTxs, err := c.memPoolModel.GetMempoolTxsByStatus(mempool.ExecutedTxStatus)
	if err != nil {
		return nil, err
	}

	if curBlock.BlockStatus > block.StatusProposing {
		if len(executedTxs) != 0 {
			return nil, errors.New("no proposing block but exist executed txs")
		}
		return curBlock, nil
	}

	for _, mempoolTx := range executedTxs {
		tx := convertMempoolTxToTx(mempoolTx)
		err = c.bc.ApplyTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	c.executedMemPoolTxs = append(c.executedMemPoolTxs, executedTxs...)
	return curBlock, nil
}

func (c *Committer) createNewBlock(curBlock *block.Block) error {
	return c.bc.BlockModel.CreateNewBlock(curBlock)
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	var now = time.Now()
	if (len(c.bc.Statedb.Txs) > 0 && now.Unix()-curBlock.CreatedAt.Unix() >= MaxCommitterInterval) ||
		len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
		return true
	}

	return false
}

func (c *Committer) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	for _, tx := range c.executedMemPoolTxs {
		tx.Status = mempool.SuccessTxStatus
	}

	blockSize := c.computeCurrentBlockSize()
	blockStates, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
	if err != nil {
		return nil, err
	}

	// Update database in a transaction.
	err = c.bc.BlockModel.CreateCompressedBlock(c.executedMemPoolTxs, blockStates)
	if err != nil {
		return nil, err
	}

	c.executedMemPoolTxs = make([]*mempool.MempoolTx, 0)
	return blockStates.Block, nil
}

func (c *Committer) computeCurrentBlockSize() int {
	var blockSize int
	for i := 0; i < len(c.optionalBlockSizes); i++ {
		if len(c.bc.Statedb.Txs) <= c.optionalBlockSizes[i] {
			blockSize = c.optionalBlockSizes[i]
			break
		}
	}
	return blockSize
}

func convertMempoolTxToTx(mempoolTx *mempool.MempoolTx) *tx.Tx {
	tx := &tx.Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		GasFee:        mempoolTx.GasFee,
		GasFeeAssetId: mempoolTx.GasFeeAssetId,
		TxStatus:      tx.StatusPending,
		NftIndex:      mempoolTx.NftIndex,
		PairIndex:     mempoolTx.PairIndex,
		AssetId:       mempoolTx.AssetId,
		TxAmount:      mempoolTx.TxAmount,
		NativeAddress: mempoolTx.NativeAddress,
		TxInfo:        mempoolTx.TxInfo,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		AccountIndex:  mempoolTx.AccountIndex,
		Nonce:         mempoolTx.Nonce,
		ExpiredAt:     mempoolTx.ExpiredAt,
	}
	return tx
}
