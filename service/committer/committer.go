package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/core"
)

const (
	MaxCommitterInterval = 60 * 1
)

var configFile = flag.String("f",
	"./etc/committer.yaml", "the config file")

type CommitterConfig struct {
	core.ChainConfig
	KeyPath struct {
		KeyTxCounts []int
	}
}

type Committer struct {
	config         *CommitterConfig
	bc             *core.BlockChain
	mempoolModel   mempool.MemPoolModel
	keyTxCounts    []int
	maxTxsPerBlock int

	pendingUpdateMempoolTxs []*mempool.MempoolTx
}

func NewCommitter(config *CommitterConfig) (*Committer, error) {
	if len(config.KeyPath.KeyTxCounts) == 0 {
		return nil, errors.New("nil key tx counts")
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
		config:                  config,
		bc:                      bc,
		mempoolModel:            mempool.NewMempoolModel(conn, config.CacheRedis, gormPointer),
		keyTxCounts:             config.KeyPath.KeyTxCounts,
		maxTxsPerBlock:          config.KeyPath.KeyTxCounts[len(config.KeyPath.KeyTxCounts)-1],
		pendingUpdateMempoolTxs: make([]*mempool.MempoolTx, 0),
	}

	go committer.loop()
	return committer, nil
}

func (c *Committer) loop() {
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
		pendingTxs, err := c.mempoolModel.GetPendingMempoolTxs()
		if err != nil {
			logx.Error("get pending transactions from mempool failed:", err)
			return
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs, err = c.mempoolModel.GetPendingMempoolTxs()
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
			tx, err = c.bc.ApplyTransaction(tx)
			if err != nil {
				mempoolTx.Status = mempool.FailTxStatus
				pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
				continue
			}
			mempoolTx.Status = mempool.ExecutedTxStatus
			pendingUpdateMempoolTxs = append(pendingUpdateMempoolTxs, mempoolTx)
		}

		err = c.bc.SyncStateCacheToRedis()
		if err != nil {
			panic("sync redis dbcache failed: " + err.Error())
		}

		err = c.mempoolModel.UpdateMempoolTxs(pendingUpdateMempoolTxs, pendingDeleteMempoolTxs)
		if err != nil {
			panic("update mempool failed: " + err.Error())
		}
		c.pendingUpdateMempoolTxs = append(c.pendingUpdateMempoolTxs, pendingUpdateMempoolTxs...)

		if c.shouldCommit(curBlock) {
			curBlock, err = c.commitNewBlock(curBlock)
			panic("commit new block failed: " + err.Error())
		}
	}
}

func (c *Committer) restoreExecutedTxs() (*block.Block, error) {
	bc := c.bc
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, err
	}
	curBlock, err := bc.BlockModel.GetBlockByBlockHeight(curHeight)
	if err != nil {
		return nil, err
	}

	executedTxs, err := c.mempoolModel.GetExecutedMempoolTxs()
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
		tx, err = c.bc.ApplyTransaction(tx)
		if err != nil {
			return nil, err
		}
		c.pendingUpdateMempoolTxs = append(c.pendingUpdateMempoolTxs, mempoolTx)
	}

	return curBlock, nil
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	var now = time.Now()
	if now.Unix()-curBlock.CreatedAt.Unix() >= MaxCommitterInterval || len(c.bc.GetPendingTxs()) >= c.maxTxsPerBlock {
		return true
	}

	return false
}

func (c *Committer) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	for _, tx := range c.pendingUpdateMempoolTxs {
		tx.Status = mempool.SuccessTxStatus
	}

	blockSize := c.computeCurrentBlockSize()
	statesToCommit, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
	if err != nil {
		return nil, err
	}

	// Update database in a transaction.
	//err = c.bc.BlockModel.CreateBlockForCommitter()

	c.pendingUpdateMempoolTxs = make([]*mempool.MempoolTx, 0)
	return statesToCommit.Block, nil
}

func (c *Committer) computeCurrentBlockSize() int {
	var blockSize int
	for i := 0; i < len(c.keyTxCounts); i++ {
		if len(c.bc.GetPendingTxs()) <= c.keyTxCounts[i] {
			blockSize = c.keyTxCounts[i]
			break
		}
	}
	return blockSize
}

func convertMempoolTxToTx(mempoolTx *mempool.MempoolTx) *tx.Tx {
	tx := &tx.Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		NativeAddress: mempoolTx.NativeAddress,
		TxInfo:        mempoolTx.TxInfo,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		Nonce:         mempoolTx.Nonce,
		ExpiredAt:     mempoolTx.ExpiredAt,
	}
	return tx
}

func main() {
	flag.Parse()
	var config CommitterConfig
	conf.MustLoad(*configFile, &config)

	_, err := NewCommitter(&config)
	if err != nil {
		logx.Error("new committer failed:", err)
		return
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
