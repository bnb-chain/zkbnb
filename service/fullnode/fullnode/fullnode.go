package fullnode

import (
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-go-sdk/client"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/tx"
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
	curHeight, err := c.bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic("get current block height failed: " + err.Error())
	}
	curHeight++

	for {
		l2Block, err := c.client.GetBlockByHeight(curHeight)
		if err != nil {
			logx.Errorf("get block failed, height: %d, err %v ", curHeight, err)
			time.Sleep(100 * time.Millisecond)
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
