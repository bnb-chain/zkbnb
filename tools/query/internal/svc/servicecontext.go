package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/tools/query/internal/config"
)

type ServiceContext struct {
	Config config.Config

	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	NftHistoryModel     nft.L2NftHistoryModel
	NftModel            nft.L2NftModel
}

func NewServiceContext(c config.Config) *ServiceContext {

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))

	return &ServiceContext{
		Config:              c,
		AccountModel:        account.NewAccountModel(db),
		AccountHistoryModel: account.NewAccountHistoryModel(db),
		NftHistoryModel:     nft.NewL2NftHistoryModel(db),
		NftModel:            nft.NewL2NftModel(db),
	}
}
