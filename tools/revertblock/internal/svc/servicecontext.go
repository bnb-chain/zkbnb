package svc

import (
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/tools/revertblock/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type ServiceContext struct {
	Config config.Config

	BlockModel      block.BlockModel
	SysConfigModel  sysconfig.SysConfigModel
	L1RollupTxModel l1rolluptx.L1RollupTxModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	masterDataSource := c.Postgres.MasterDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(masterDataSource)},
	}))

	return &ServiceContext{
		Config:          c,
		BlockModel:      block.NewBlockModel(db),
		SysConfigModel:  sysconfig.NewSysConfigModel(db),
		L1RollupTxModel: l1rolluptx.NewL1RollupTxModel(db),
	}
}
