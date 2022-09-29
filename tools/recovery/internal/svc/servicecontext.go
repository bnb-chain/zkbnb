package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/tools/recovery/internal/config"
)

type ServiceContext struct {
	Config config.Config

	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	NftHistoryModel     nft.L2NftHistoryModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	return &ServiceContext{
		Config:              c,
		AccountModel:        account.NewAccountModel(db),
		AccountHistoryModel: account.NewAccountHistoryModel(db),
		NftHistoryModel:     nft.NewL2NftHistoryModel(db),
	}
}
