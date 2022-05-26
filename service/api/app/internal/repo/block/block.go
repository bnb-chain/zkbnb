package block

import (
	"log"
	"strings"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"

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
	Txs        []*tx.Tx `gorm:"foreignkey:BlockId"`
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: GetExecutedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *block) GetExecutedBlocksCount() (count int64, err error) {
	result, err := m.cache.Get(cacheBlockExecutedCountPrefix, count,
		multcache.SqlQueryCount, m.db, m.table,
		"block_status = ? and deleted_at is NULL", StatusExecuted)
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
	}
	return count, nil
}

/*
	Func: GetCommitedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *block) GetCommitedBlocksCount() (count int64, err error) {
	result, err := m.cache.Get(cacheBlockCommittedCountPrefix, count,
		multcache.SqlQueryCount, m.db, m.table,
		"block_status >= ? and deleted_at is NULL", StatusCommitted)
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
	}
	return count, nil
}

/*
	Func: GetBlockByBlockHeight
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *block) GetBlockByBlockHeight(blockHeight int64) (block *BlockInfo, err error) {
	var (
		blockForeignKeyColumn = `BlockDetails`
		txForeignKeyColumn    = `Txs`
	)
	dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	err = m.db.Model(&block).Association(blockForeignKeyColumn).Find(&block.BlockDetails)
	if err != nil {
		return nil, err
	}
	err = m.db.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	if err != nil {
		return nil, err
	}
	return block, nil
}
