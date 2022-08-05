package txdetail

import (
	"context"
	"sort"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetTxsTotalCountByAccountIndex(ctx context.Context, accountIndex int64) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Select("tx_id").
			Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
		if dbTx.Error != nil {
			logx.Errorf("fail to get tx count by account: %d, error: %s", accountIndex, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, nil
		}
		return &count, nil
	}
	var countType int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxsCount(), &countType, multcache.TxCountTtl, f)
	if err != nil {
		return 0, err
	}
	count, _ := value.(*int64)
	return *count, nil
}

func (m *model) GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error) {
	result := make([]*table.TxDetail, 0)
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&result)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx details by account: %d, error: %s", accountIndex, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return result, nil
}

func (m *model) GetTxIdsByAccountIndex(ctx context.Context, accountIndex int64) ([]int64, error) {
	txIds := make([]int64, 0)
	dbTx := m.db.Table(m.table).Select("tx_id").Where("account_index = ?", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		logx.Errorf("fail to get tx ids by account: %d, error: %s", accountIndex, dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	sort.Slice(txIds, func(i, j int) bool {
		return txIds[i] > txIds[j]
	})
	return txIds, nil
}

func (m *model) GetDauInTxDetail(ctx context.Context, data string) (int64, error) {
	var (
		from time.Time
		to   time.Time
	)
	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)
	switch data {
	case "yesterday":
		from = today.Add(-24 * time.Hour)
		to = today
	case "today":
		from = today
		to = now
	}
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Raw("SELECT account_index FROM tx_detail WHERE created_at BETWEEN ? AND ? AND account_index != -1 GROUP BY account_index", from, to).Count(&count)
		if dbTx.Error != nil {
			logx.Errorf("fail to get dau by time range: %d-%d, error: %s", from.Unix(), to.Unix(), dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, nil
		}
		return &count, nil
	}
	var countType int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxCountByTimeRange(data), &countType, multcache.DauTtl, f)
	if err != nil {
		return 0, err
	}
	count, _ := value.(*int64)
	return *count, nil
}
