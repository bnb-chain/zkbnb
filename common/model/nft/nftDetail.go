package nft

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	AccountL2NftDetailModel interface {
		CreateAccountL2NftDetailTable() error
		DropAccountL2NftDetailTable() error
	}
	defaultAccountL2NftDetailModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountL2NftDetail struct {
		gorm.Model
		NftId          int64
		NftContentHash string
		Url            string
		Name           string
		Introduction   string
		Attributes     string
		NftL1Url       string
	}
)

func NewAccountL2NftDetailModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountL2NftDetailModel {
	return &defaultAccountL2NftDetailModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AccountL2NftDetailTableName,
		DB:         db,
	}
}

func (*AccountL2NftDetail) TableName() string {
	return AccountL2NftDetailTableName
}

/*
	Func: CreateAccountL2NftDetailTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultAccountL2NftDetailModel) CreateAccountL2NftDetailTable() error {
	return m.DB.AutoMigrate(AccountL2NftDetail{})
}

/*
	Func: DropAccountL2NftDetailTable
	Params:
	Return: err error
	Description: drop accountnft table
*/
func (m *defaultAccountL2NftDetailModel) DropAccountL2NftDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
