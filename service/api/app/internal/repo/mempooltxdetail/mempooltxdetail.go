package mempooltxdetail

import (
	"context"
	table "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetMempoolTxs
	Params: offset uint64, limit uint64
	Return: mempoolTx []*mempoolModel.MempoolTx, err error
	Description: query txs from db that sit in the range
*/
func (m *model) GetMemPoolTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.MempoolTxDetail, error) {
	result := make([]*table.MempoolTxDetail, 0)
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Order("created_at").Find(&result)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.RepoErrDataNotExist
	}
	return result, nil
}
