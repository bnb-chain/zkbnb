package nft

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L2NftExchangeModel interface {
		CreateL2NftExchangeTable() error
		DropL2NftExchangeTable() error
	}
	defaultL2NftExchangeModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L2NftExchange struct {
		gorm.Model
		BuyerAccountIndex int64
		OwnerAccountIndex int64
		NftIndex          int64
		AssetId           int64
		AssetAmount       int64
	}
)

func NewL2NftExchangeModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L2NftExchangeModel {
	return &defaultL2NftExchangeModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      L2NftExchangeTableName,
		DB:         db,
	}
}

func (*L2NftExchange) TableName() string {
	return L2NftExchangeTableName
}

/*
	Func: CreateL2NftExchangeTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultL2NftExchangeModel) CreateL2NftExchangeTable() error {
	return m.DB.AutoMigrate(L2NftExchange{})
}

/*
	Func: DropAccountNFTTable
	Params:
	Return: err error
	Description: drop accountnft table
*/
func (m *defaultL2NftExchangeModel) DropL2NftExchangeTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
