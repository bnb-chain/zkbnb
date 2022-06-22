package block

import (
	"fmt"
	"strings"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"

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
		return nil, ErrDataNotExistInSQL
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
	dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	for _, block := range blocks {
		err = m.db.Model(&block).Association(`Txs`).Find(&block.Txs)
		if err != nil {
			return nil, err
		}
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
