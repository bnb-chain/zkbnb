package block

import (
	"fmt"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/errcode"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type block struct {
	table     string
	db        *gorm.DB
	cache     multcache.MultCache
	redisConn *redis.Redis
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
		return nil, errcode.ErrDataNotExist
	}
	return
}

func (m *block) GetCommitedBlocksCount() (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusCommitted).Count(&count).Error
	return
}

func (m *block) GetVerifiedBlocksCount() (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", table.StatusVerifiedAndExecuted).Count(&count).Error
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
