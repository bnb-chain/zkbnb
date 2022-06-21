package block

import (
	"encoding/json"
	"fmt"
	"strings"

	table "github.com/bnb-chain/zkbas/common/model/block"
	tableTx "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/pkg/multcache"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/builderx"
	"gorm.io/gorm"
)

var (
	blockFieldNames          = builderx.RawFieldNames(&block{})
	blockRows                = strings.Join(blockFieldNames, ",")
	blockRowsExpectAutoSet   = strings.Join(stringx.Remove(blockFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	blockRowsWithPlaceHolder = strings.Join(stringx.Remove(blockFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheBlockIdPrefix              = "cache::block:id:"
	cacheBlockBlockCommitmentPrefix = "cache::block:blockCommitment:"
	cacheBlockHeightPrefix          = "cache::block:blockHeight:"
	CacheBlockStatusPrefix          = "cache::block:blockStatus:"
	cacheBlockListLimitPrefix       = "cache::block:blockList:"
	cacheBlockCommittedCountPrefix  = "cache::block:committed_count"
	cacheBlockVerifiedCountPrefix   = "cache::block:verified_count"
	cacheBlockExecutedCountPrefix   = "cache::block:executed_count"
)

type block struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
	redisConn  *redis.Redis
}

/*
	Func: GetBlockByBlockHeight
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *block) GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error) {
	dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", dbTx.Error)
	} else if dbTx.RowsAffected == 0 {
		return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", ErrNotFound)
	}
	return
}

func (m *block) GetCommitedBlocksCount() (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusCommitted).Count(&count).Error
	return
}

func (m *block) GetExecutedBlocksCount() (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusExecuted).Count(&count).Error
	return
}

func (m *block) GetBlockWithTxsByCommitment(blockCommitment string) (block *table.Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.db.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(&block)
	if dbTx.Error != nil {
		logx.Error("[block.GetBlockByCommitment] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlockByCommitment] Get Block Error")
		return nil, ErrNotFound
	}
	err = m.db.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	if err != nil {
		logx.Error("[block.GetBlockByCommitment] Get Associate Txs Error")
		return nil, err
	}
	return block, nil
}

func (m *block) GetBlockWithTxsByBlockHeight(blockHeight int64) (block *table.Block, err error) {
	txForeignKeyColumn := `Txs`
	dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", dbTx.Error)
	} else if dbTx.RowsAffected == 0 {
		return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", ErrNotFound)
	}
	err = m.db.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	if err != nil {
		return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", err)
	}
	return block, nil
}

func (m *block) GetBlocksList(limit int64, offset int64) (blocks []*table.Block, err error) {
	var (
		//blockForeignKeyColumn = `BlockDetails`
		txForeignKeyColumn = `Txs`
	)
	key := fmt.Sprintf("%s%v:%v", cacheBlockListLimitPrefix, limit, offset)
	cacheBlockListLimitVal, err := m.redisConn.Get(key)

	if err != nil {
		errInfo := fmt.Sprintf("[block.GetBlocksList] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return nil, err
	} else if cacheBlockListLimitVal == "" {
		dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
		if dbTx.Error != nil {
			logx.Error("[block.GetBlocksList] %s", dbTx.Error)
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			logx.Error("[block.GetBlocksList] Get Blocks Error")
			return nil, ErrNotFound
		}

		for _, block := range blocks {
			cacheBlockIdKey := fmt.Sprintf("%s%v", cacheBlockIdPrefix, block.ID)
			cacheBlockIdVal, err := m.redisConn.Get(cacheBlockIdKey)
			if err != nil {
				errInfo := fmt.Sprintf("[block.GetBlocksList] Get Redis Error: %s, key:%s", err.Error(), key)
				logx.Errorf(errInfo)
				return nil, err
			} else if cacheBlockIdVal == "" {
				/*
					err = m.DB.Model(&block).Association(blockForeignKeyColumn).Find(&block.BlockDetails)
					if err != nil {
						logx.Error("[block.GetBlocksList] Get Associate BlockDetails Error")
						return nil, err
					}
				*/
				txLength := m.db.Model(&block).Association(txForeignKeyColumn).Count()
				block.Txs = make([]*tableTx.Tx, txLength)

				// json string
				jsonString, err := json.Marshal(block)
				if err != nil {
					logx.Errorf("[block.GetBlocksList] json.Marshal Error: %s, value: %v", block)
					return nil, err
				}
				// todo
				err = m.redisConn.Setex(key, string(jsonString), 60)
				if err != nil {
					logx.Errorf("[block.GetBlocksList] redis set error: %s", err.Error())
					return nil, err
				}
			} else {
				// json string unmarshal
				var (
					nBlock *table.Block
				)
				err = json.Unmarshal([]byte(cacheBlockIdVal), &nBlock)
				if err != nil {
					logx.Errorf("[tblock.GetBlocksList] json.Unmarshal error: %s, value : %s", err.Error(), cacheBlockIdVal)
					return nil, err
				}
				block = nBlock
			}
		}
		// json string
		jsonString, err := json.Marshal(blocks)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] json.Marshal Error: %s, value: %v", err.Error(), blocks)
			return nil, err
		}
		// todo
		err = m.redisConn.Setex(key, string(jsonString), 30)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] redis set error: %s", err.Error())
			return nil, err
		}

	} else {
		// json string unmarshal
		var (
			nBlocks []*table.Block
		)
		err = json.Unmarshal([]byte(cacheBlockListLimitVal), &nBlocks)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] json.Unmarshal error: %s, value : %s", err.Error(), cacheBlockListLimitVal)
			return nil, err
		}
		blocks = nBlocks
	}

	return blocks, nil
}

func (m *block) GetBlocksTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Error("[block.GetBlocksTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[block.GetBlocksTotalCount] No Blocks in Block Table")
		return 0, nil
	}
	return count, nil
}
