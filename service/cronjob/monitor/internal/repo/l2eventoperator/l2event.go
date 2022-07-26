package l2eventoperator

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) UpdateL2Events(pendingUpdateL2Events []*l2TxEventMonitor.L2TxEventMonitor) (err error) {
	for _, pendingUpdateL2Event := range pendingUpdateL2Events {
		dbTx := m.db.Table(m.table).Where("id = ?", pendingUpdateL2Event.ID).Select("*").Updates(&pendingUpdateL2Event)
		if dbTx.Error != nil {
			logx.Errorf("[CreateMempoolAndActiveAccount] unable to update l2 tx event: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[CreateMempoolAndActiveAccount] invalid l2 tx event")
			return errors.New("[CreateMempoolAndActiveAccount] invalid l2 tx event")
		}
	}
	return nil
}
