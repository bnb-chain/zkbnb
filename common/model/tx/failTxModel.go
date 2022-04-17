package tx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyFailTxIdPrefix = "cache:zecrey:failTx:id:"
)

type (
	FailTxModel interface {
		CreateFailTxTable() error
		DropFailTxTable() error
		CreateFailTx(failTx *FailTx) error
	}

	defaultFailTxModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	FailTx struct {
		gorm.Model
		TxHash        string `gorm:"uniqueIndex"`
		TxType        int64
		GasFee        int64
		GasFeeAssetId int64
		TxStatus      int64
		AssetAId      int64
		AssetBId      int64
		TxAmount      int64
		NativeAddress string
		ChainId       int64
		TxInfo        string
		ExtraInfo     string
		Memo          string
	}
)

func NewFailTxModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) FailTxModel {
	return &defaultFailTxModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `fail_tx`,
		DB:         db,
	}
}

func (*FailTx) TableName() string {
	return `fail_tx`
}

/*
	Func: CreateFailTxTable
	Params:
	Return: err error
	Description: create tx fail table
*/
func (m *defaultFailTxModel) CreateFailTxTable() error {
	return m.DB.AutoMigrate(FailTx{})
}

/*
	Func: DropFailTxTable
	Params:
	Return: err error
	Description: drop tx fail table
*/
func (m *defaultFailTxModel) DropFailTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateFailTx
	Params: failTx *FailTx
	Return: err error
	Description: create fail tx
*/
func (m *defaultFailTxModel) CreateFailTx(failTx *FailTx) error {
	dbTx := m.DB.Table(m.table).Create(failTx)
	if dbTx.Error != nil {
		logx.Error("[tx.CreateFailTx] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[tx.CreateFailTx] Create Invalid Fail Tx")
		return ErrInvalidFailTx
	}
	return nil
}
