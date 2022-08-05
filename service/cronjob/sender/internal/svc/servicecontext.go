package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/internal/config"
)

type ServiceContext struct {
	Config config.Config

	BlockModel          block.BlockModel
	BlockForCommitModel blockForCommit.BlockForCommitModel
	L1TxSenderModel     l1TxSender.L1TxSenderModel
	SysConfigModel      sysconfig.SysconfigModel
	ProofSenderModel    proofSender.ProofSenderModel
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	return &ServiceContext{
		Config:              c,
		BlockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		BlockForCommitModel: blockForCommit.NewBlockForCommitModel(conn, c.CacheRedis, gormPointer),
		L1TxSenderModel:     l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		ProofSenderModel:    proofSender.NewProofSenderModel(gormPointer),
	}
}
