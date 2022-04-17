package l2TxEventMonitor

import (
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2TxEventMonitorModel interface {
		CreateL2TxEventMonitorTable() error
		DropL2TxEventMonitorTable() error
		CreateL2TxEventMonitor(tx *L2TxEventMonitor) (bool, error)
		CreateL2TxEventMonitorsInBatches(l2TxEventMonitors []*L2TxEventMonitor) (rowsAffected int64, err error)
		GetL2TxEventMonitorsByStatus(status int) (txs []*L2TxEventMonitor, err error)
		GetL2TxEventMonitorsByAccountName(accountName string) (txs []*L2TxEventMonitor, err error)
		GetL2TxEventMonitorsByTxType(txType uint8) (txs []*L2TxEventMonitor, err error)
	}

	defaultL2TxEventMonitorModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2TxEventMonitor struct {
		gorm.Model
		// tx type
		TxType uint8 `gorm:"index"`
		// chain id
		ChainId uint8
		// asset id
		AssetId uint32
		// related tx hash
		L1TxHash string
		// related block height
		L1BlockHeight int64
		// account name
		AccountName string
		// native address
		NativeAddress string
		// layer-2 amount
		Amount string
		// balance delta
		BalanceDelta string
		// status
		Status int
	}
)

func (*L2TxEventMonitor) TableName() string {
	return TableName
}

func NewL2TxEventMonitorModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2TxEventMonitorModel {
	return &defaultL2TxEventMonitorModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL2TxEventMonitorTable
	Params:
	Return: err error
	Description: create l2 tx event monitor table
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitorTable() error {
	return m.DB.AutoMigrate(L2TxEventMonitor{})
}

/*
	Func: DropL2TxEventMonitorTable
	Params:
	Return: err error
	Description: drop l2 tx event monitor table
*/
func (m *defaultL2TxEventMonitorModel) DropL2TxEventMonitorTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL2TxEventMonitor
	Params: asset *L2TxEventMonitor
	Return: bool, error
	Description: create L2TxEventMonitor tx
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitor(tx *L2TxEventMonitor) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL2TxEventMonitor] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL2TxEventMonitor := errors.New("invalid l2TxEventMonitor")
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL2TxEventMonitor] %s", ErrInvalidL2TxEventMonitor)
		logx.Error(err)
		return false, ErrInvalidL2TxEventMonitor
	}
	return true, nil
}

/*
	Func: CreateL2TxEventMonitorsInBatches
	Params: []*L2TxEventMonitor
	Return: rowsAffected int64, err error
	Description: create L2TxEventMonitor batches
*/
func (m *defaultL2TxEventMonitorModel) CreateL2TxEventMonitorsInBatches(l2TxEventMonitors []*L2TxEventMonitor) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(l2TxEventMonitors, len(l2TxEventMonitors))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.CreateL1AssetsMonitorInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	GetL2TxEventMonitors: get all L2TxEventMonitors
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitors() (txs []*L2TxEventMonitor, err error) {
	dbTx := m.DB.Table(m.table).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitors] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitors] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, dbTx.Error
}

/*
	Func: GetPendingL2TxEventMonitors
	Return: txs []*L2TxEventMonitor, err error
	Description: get pending l2TxEventMonitors
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsByStatus(status int) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("status = ?", status).Find(&txs).Order("create_at")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByStatus] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByStatus] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL2TxEventMonitorsByAccountName
	Return: txs []*L2TxEventMonitor, err error
	Description: get l2TxEventMonitors by account name
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsByAccountName(accountName string) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("account_name = ?", accountName).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByAccountName] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByAccountName] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}

/*
	Func: GetL2TxEventMonitorsByTxType
	Return: txs []*L2TxEventMonitor, err error
	Description: get l2TxEventMonitors by tx type
*/
func (m *defaultL2TxEventMonitorModel) GetL2TxEventMonitorsByTxType(txType uint8) (txs []*L2TxEventMonitor, err error) {
	// todo order id
	dbTx := m.DB.Table(m.table).Where("tx_type = ?", txType).Find(&txs).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByTxType] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l2TxEventMonitor.GetL2TxEventMonitorsByTxType] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return txs, nil
}
