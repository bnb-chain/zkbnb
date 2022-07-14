package block

import (
	"context"
	"fmt"
	"strconv"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/errcode"

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
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockByBlockHeight+strconv.FormatInt(blockHeight, 10), _block, 1, f)
	if err != nil {
		return nil, err
	}
	_block, _ = value.(*table.Block)
	return _block, nil

}

func (m *block) GetCommitedBlocksCount(ctx context.Context) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", table.StatusCommitted).Count(&count)
		if dbTx.Error != nil {
			return nil, fmt.Errorf("[GetCommitedBlocksCount]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist
		}
		return &count, nil
	}
	var count int64
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetCommittedBlocksCount, &count, 5, f)
	if err != nil {
		return count, err
	}
	count1, _ := value.(*int64)
	return *count1, nil
}

func (m *block) GetVerifiedBlocksCount(ctx context.Context) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Where("block_status = ? and deleted_at is NULL", table.StatusVerifiedAndExecuted).Count(&count)
		if dbTx.Error != nil {
			return nil, fmt.Errorf("[GetCommitedBlocksCount]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist
		}
		return &count, nil
	}
	var count int64
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetVerifiedBlocksCount, &count, 5, f)
	if err != nil {
		return count, err
	}
	count1, _ := value.(*int64)
	return *count1, nil

}

func (m *block) GetBlockWithTxsByCommitment(ctx context.Context, blockCommitment string) (*table.Block, error) {
	f := func() (interface{}, error) {
		txForeignKeyColumn := `Txs`
		_block := &table.Block{}
		dbTx := m.db.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(_block)
		if dbTx.Error != nil {
			return nil, fmt.Errorf("[GetBlockWithTxsByCommitment] %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist
		}
		err := m.db.Model(&_block).Association(txForeignKeyColumn).Find(&_block.Txs)
		if err != nil {
			return nil, fmt.Errorf("[GetBlockWithTxsByCommitment] Get Associate Txs Error: %v", ErrNotFound)
		}
		return _block, nil
	}
	_block := &table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockBlockCommitment+blockCommitment, _block, 1, f)
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
			return nil, fmt.Errorf("[GetBlockWithTxsByBlockHeight]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return nil, fmt.Errorf("[GetBlockWithTxsByBlockHeight]: %v", ErrNotFound)
		}
		err := m.db.Model(&_block).Association(txForeignKeyColumn).Find(&_block.Txs)
		if err != nil {
			return nil, fmt.Errorf("[GetBlockWithTxsByBlockHeight]: %v", err)
		}
		return _block, nil
	}
	_block := &table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockWithTxHeight+strconv.FormatInt(blockHeight, 10), _block, 1, f)
	if err != nil {
		return nil, err
	}
	_block, _ = value.(*table.Block)
	return _block, nil

}

func (m *block) GetBlocksList(ctx context.Context, limit int64, offset int64) ([]*table.Block, error) {
	f := func() (interface{}, error) {
		blockList := []*table.Block{}
		dbTx := m.db.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blockList)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, ErrNotFound
		}
		for _, _block := range blockList {
			err := m.db.Model(&_block).Association(`Txs`).Find(&_block.Txs)
			if err != nil {
				return nil, err
			}
		}
		return &blockList, nil
	}
	blockList := []*table.Block{}
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlockList+strconv.FormatInt(limit, 10)+strconv.FormatInt(offset, 10), &blockList, 1, f)
	if err != nil {
		return nil, err
	}
	blockList1, ok := value.(*[]*table.Block)
	if !ok {
		return nil, fmt.Errorf("[GetBlocksList] ErrConvertFail")
	}
	return *blockList1, nil
}

func (m *block) GetBlocksTotalCount(ctx context.Context) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Where("deleted_at is NULL").Count(&count)
		if dbTx.Error != nil {
			return 0, fmt.Errorf("[GetBlocksTotalCount]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return 0, ErrNotFound
		}
		return &count, nil
	}
	var count int64
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetBlocksTotalCount, &count, 1, f)
	if err != nil {
		return count, err
	}
	count1, ok := value.(*int64)
	if !ok {
		return 0, fmt.Errorf("[GetBlocksTotalCount] ErrConvertFail")
	}
	return *count1, nil
}
