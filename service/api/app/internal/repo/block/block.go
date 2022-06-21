package block

import (
	table "github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/pkg/multcache"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

type block struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
}

/*
	Func: GetBlockByBlockHeight
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *block) GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error) {
	txForeignKeyColumn := `Txs`
	dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	err = m.db.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	if err != nil {
		return nil, err
	}
	return block, nil
}
