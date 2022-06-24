package block

import (
	"context"
	"fmt"
	"strconv"

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
func (m *block) GetBlockByBlockHeight(ctx context.Context, blockHeight int64) (*table.Block, error) {
	f := func() (interface{}, error) {
		_block := &table.Block{}
		dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(_block)
		if dbTx.Error != nil {
			return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist
		}
		return _block, nil
	}
	_block := &table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockByBlockHeight+strconv.FormatInt(blockHeight, 10), _block, 10, f)
	if err != nil {
		return nil, err
	}
	_block, _ = value.(*table.Block)
	return _block, nil

}

func (m *block) GetCommitedBlocksCount(_ context.Context) (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusCommitted).Count(&count).Error
	return
}

func (m *block) GetVerifiedBlocksCount(_ context.Context) (count int64, err error) {
	err = m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", table.StatusVerifiedAndExecuted).Count(&count).Error
	return
}

func (m *block) GetBlockWithTxsByCommitment(ctx context.Context, blockCommitment string) (*table.Block, error) {
	f := func() (interface{}, error) {
		txForeignKeyColumn := `Txs`
		_block := &table.Block{}
		dbTx := m.db.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(_block)
		if dbTx.Error != nil {
			logx.Error("[block.GetBlockByCommitment] %s", dbTx.Error)
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			logx.Error("[block.GetBlockByCommitment] Get Block Error")
			return nil, ErrNotFound
		}
		err := m.db.Model(&_block).Association(txForeignKeyColumn).Find(&_block.Txs)
		if err != nil {
			logx.Error("[block.GetBlockByCommitment] Get Associate Txs Error")
			return nil, err
		}
		return _block, nil
	}
	_block := &table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockBlockCommitment+blockCommitment, _block, 10, f)
	if err != nil {
		return nil, err
	}
	_block, _ = value.(*table.Block)
	return _block, nil

}

func (m *block) GetBlockWithTxsByBlockHeight(ctx context.Context, blockHeight int64) (*table.Block, error) {
	f := func() (interface{}, error) {
		txForeignKeyColumn := `Txs`
		_block := &table.Block{}
		dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(_block)
		if dbTx.Error != nil {
			return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", ErrNotFound)
		}
		err := m.db.Model(&_block).Association(txForeignKeyColumn).Find(&_block.Txs)
		if err != nil {
			return nil, fmt.Errorf("[GetBlockByBlockHeight]: %v", err)
		}
		return _block, nil
	}
	_block := &table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockWithTxHeight+strconv.FormatInt(blockHeight, 10), _block, 10, f)
	if err != nil {
		return nil, err
	}
	_block, _ = value.(*table.Block)
	return _block, nil

}

func (m *block) GetBlocksList(ctx context.Context, limit int64, offset int64) ([]*table.Block, error) {
	f := func() (interface{}, error) {
		blockList := &[]*table.Block{}
		dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blockList)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, ErrNotFound
		}
		for _, _block := range *blockList {
			err := m.db.Model(&_block).Association(`Txs`).Find(&_block.Txs)
			if err != nil {
				return nil, err
			}
		}
		return blockList, nil
	}
	blockList := &[]*table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockList+strconv.FormatInt(limit, 10)+strconv.FormatInt(offset, 10), blockList, 10, f)
	if err != nil {
		return nil, err
	}
	blockList, ok := value.(*[]*table.Block)
	fmt.Println(blockList, ok)
	return *blockList, nil
}

func (m *block) GetBlocksTotalCount(_ context.Context) (count int64, err error) {
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
