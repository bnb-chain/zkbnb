package mempooltxdetail

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
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
		logx.Errorf("fail to get mempool tx by account: %d, error: %s", accountIndex, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return result, nil
}
