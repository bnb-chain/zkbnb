package tx

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyTxDetailIdPrefix = "cache:zecrey:txDetail:id:"
)

type (
	TxDetailModel interface {
		CreateTxDetailTable() error
		DropTxDetailTable() error
		GetTxDetailsByAccountName(name string) (txDetails []*TxDetail, err error)
		UpdateTxDetail(detail *TxDetail) error
	}

	defaultTxDetailModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	TxDetail struct {
		gorm.Model
		TxId         int64 `gorm:"index"`
		AssetId      int64
		AssetType    int64
		AccountIndex int64 `gorm:"index"`
		AccountName  string
		Balance      string
		BalanceDelta string
	}
)

func NewTxDetailModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) TxDetailModel {
	return &defaultTxDetailModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TxDetailTableName,
		DB:         db,
	}
}

func (*TxDetail) TableName() string {
	return TxDetailTableName
}

/*
	Func: CreateTxDetailTable
	Params:
	Return: err error
	Description: create tx detail table
*/
func (m *defaultTxDetailModel) CreateTxDetailTable() error {
	return m.DB.AutoMigrate(TxDetail{})
}

/*
	Func: DropTxDetailTable
	Params:
	Return: err error
	Description: drop tx detail table
*/
func (m *defaultTxDetailModel) DropTxDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetTxDetailsByAccountName
	Params: name string
	Return: txDetails []*TxDetail, err error
	Description: GetTxDetailsByAccountName
*/
func (m *defaultTxDetailModel) GetTxDetailsByAccountName(name string) (txDetails []*TxDetail, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", name).Find(&txDetails)
	if dbTx.Error != nil {
		if dbTx.Error == ErrNotFound {
			return nil, nil
		} else {
			return nil, dbTx.Error
		}
	} else {
		return txDetails, nil
	}
}

func (m *defaultTxDetailModel) UpdateTxDetail(detail *TxDetail) error {
	dbTx := m.DB.Save(&detail)
	if dbTx.Error != nil {
		if dbTx.Error == ErrNotFound {
			return nil
		} else {
			return dbTx.Error
		}
	} else {
		return nil
	}
}
