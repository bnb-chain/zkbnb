package l1amount

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreyL1AmountIdPrefix   = "cache:zecrey:l1amount:id:"
	cacheZecreyL1AmountNamePrefix = "cache:zecrey:l1amount:blockHeight:"
)

type (
	L1AmountModel interface {
		CreateL1AmountTable() error
		DropL1AmountTable() error
		GetLatestL1AmountInfo() (amountInfos []*L1Amount, err error)
		GetL1AmountById(id uint) (amountInfo *L1Amount, err error)
		CreateL1Amount(l1amount *L1Amount) error
		CreateL1AmountInBatches(l1amounts []*L1Amount) error
		UpdateL1Amount(l1amount *L1Amount) error
	}

	defaultL1AmountModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1Amount struct {
		gorm.Model
		AssetId     uint32 `gorm:"index"`
		BlockHeight int64  `gorm:"index"`
		TotalAmount int64
	}
)

func NewL1AmountModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L1AmountModel {
	return &defaultL1AmountModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `l1_amount`,
		DB:         db,
	}
}

func (*L1Amount) TableName() string {
	return `l1_amount`
}

/*
	Func: CreateL1AmountTable
	Params:
	Return: err error
	Description: create L1Amount table
*/
func (m *defaultL1AmountModel) CreateL1AmountTable() error {
	return m.DB.AutoMigrate(L1Amount{})
}

/*
	Func: DropL1AmountTable
	Params:
	Return: err error
	Description: drop L1Amount table
*/
func (m *defaultL1AmountModel) DropL1AmountTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL1Amount
	Params: l1amount *L1Amount
	Return: error
	Description: Insert New L1Amount
*/

func (m *defaultL1AmountModel) CreateL1Amount(l1amount *L1Amount) error {
	dbTx := m.DB.Table(m.table).Create(l1amount)
	if dbTx.Error != nil {
		logx.Errorf("[l1amount.CreateL1Amount] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Errorf("[l1amount.CreateL1Amount] Delete Invalid Mempool Tx")
		return ErrInvalidL1Amount
	}
	return nil
}

func (m *defaultL1AmountModel) CreateL1AmountInBatches(l1amounts []*L1Amount) error {
	dbTx := m.DB.Table(m.table).CreateInBatches(l1amounts, len(l1amounts))
	if dbTx.Error != nil {
		logx.Errorf("[l1amount.CreateL1Amount] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 || dbTx.RowsAffected != int64(len(l1amounts)) {
		logx.Errorf("[l1amount.CreateL1Amount] invalid l1 amount")
		return ErrInvalidL1Amount
	}
	return nil
}

/*
	Func: UpdateL1Amount
	Params: l1amount *L1Amount
	Return: err error
	Description: update l1amount
*/
func (m *defaultL1AmountModel) UpdateL1Amount(l1amount *L1Amount) error {
	dbTx := m.DB.Table(m.table).
		Where("asset_id = ? and block_height = ?", l1amount.AssetId, l1amount.BlockHeight).
		Select(TotalAmountColumn).
		Updates(l1amount)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1amount.UpdateL1Amount] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1amount.UpdateL1Amount] %s", ErrNotFound)
		logx.Error(err)
		return ErrNotFound
	}
	return nil
}

func (m *defaultL1AmountModel) GetLatestL1AmountInfo() (amountInfos []*L1Amount, err error) {
	dbTx := m.DB.Debug().Table(m.table).Raw("select * from l1_amount where id in (select max(id) from l1_amount GROUP BY chain_id, asset_id) ORDER BY chain_id, asset_id").Find(&amountInfos)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestL1AmountInfo] unable to get latest l1 amount info: %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetLatestL1AmountInfo] error not found: %s", ErrNotFound)
		return nil, ErrNotFound
	}
	return amountInfos, nil
}

func (m *defaultL1AmountModel) GetL1AmountById(id uint) (amountInfo *L1Amount, err error) {
	m.DB.Table(m.table).Where("id = ?", id).Find(&amountInfo)
	return amountInfo, nil
}
