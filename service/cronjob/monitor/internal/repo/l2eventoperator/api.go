package l2eventoperator

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
)

type Model interface {
	UpdateL2Events(pendingUpdateL2Events []*table.L2TxEventMonitor) (err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: "l2_tx_event_monitor",
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
