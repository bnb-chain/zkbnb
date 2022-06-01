package svc

import (
	"github.com/zecrey-labs/zecrey/common/model/block"
	"github.com/zecrey-labs/zecrey/common/model/blockForProver"
	"github.com/zecrey-labs/zecrey/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey/common/model/proofSender"
	"github.com/zecrey-labs/zecrey/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey/service/cronjob/sender/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config

	BlockModel          block.BlockModel
	BlockDetailModel    block.BlockDetailModel
	BlockForProverModel blockForProver.BlockForProverModel
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
		Config: c,

		BlockModel:       block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		BlockDetailModel: block.NewBlockDetailModel(conn, c.CacheRedis, gormPointer),
		L1TxSenderModel:  l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:   sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		ProofSenderModel: proofSender.NewProofSenderModel(gormPointer),
	}
}
