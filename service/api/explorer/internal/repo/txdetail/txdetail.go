package txdetail

import (
	"context"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/errcode"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error) {
	result := make([]*table.TxDetail, 0)
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&result)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errcode.ErrDataNotExist
	}
	return result, nil
}
