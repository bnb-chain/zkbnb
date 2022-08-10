package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/core"
)

var configFile = flag.String("f",
	"./etc/committer.yaml", "the config file")

type Committer struct {
	config *core.ChainConfig
	bc     *core.BlockChain
}

func NewCommitter(config *core.ChainConfig) (*Committer, error) {
	bc, err := core.NewBlockChain(config, "committer")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	committer := &Committer{
		config: config,
		bc:     bc,
	}

	go committer.loop()
	return committer, nil
}

func (c *Committer) loop() {
	curBlock, stateCache, err := c.restoreExecutedTxs()
	if err != nil {
		logx.Error("restore executed transactions failed:", err)
	}

	for {
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock = c.proposeNewBlock()
			stateCache = core.NewStateCache(curBlock.BlockHeight)
		}

		// Read pending transactions from mempool_tx table.
		pendingTxs, err := c.bc.MempoolModel.GetMempoolTxsListForCommitter()
		if err != nil {
			logx.Error("get pending transactions from mempool failed:", err)
			return
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit() {
				break
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs, err = c.bc.MempoolModel.GetMempoolTxsListForCommitter()
			if err != nil {
				logx.Error("get pending transactions from mempool failed:", err)
				return
			}
		}

		for _, mempoolTx := range pendingTxs {
			if c.shouldCommit() {
				break
			}

			// Convert mempoolTx to tx.
			tx := convertMempoolTxToTx(mempoolTx)
			tx, stateCache, err = c.bc.ApplyTransaction(tx, stateCache)
			if err != nil {
				logx.Error("apply tx failed:", err)
				return
			}
		}

		// TODO: update executed transaction and update cache.

		if c.shouldCommit() {
			c.commitNewBlock()
		}
	}
}

func (c *Committer) restoreExecutedTxs() (*block.Block, *core.StateCache, error) {
	return nil, nil, nil
}

func (c *Committer) shouldCommit() bool {
	return false
}

func (c *Committer) proposeNewBlock() *block.Block {
	return nil
}

func (c *Committer) commitNewBlock() {

}

func convertMempoolTxToTx(mempoolTx *mempool.MempoolTx) *tx.Tx {
	tx := &tx.Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		GasFee:        mempoolTx.GasFee,
		GasFeeAssetId: mempoolTx.GasFeeAssetId,
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

func main() {
	flag.Parse()
	var chainConfig core.ChainConfig
	conf.MustLoad(*configFile, &chainConfig)

	_, err := NewCommitter(&chainConfig)
	if err != nil {
		logx.Error("new committer failed:", err)
		return
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
