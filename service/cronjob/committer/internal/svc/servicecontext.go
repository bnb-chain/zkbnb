package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/service/cronjob/committer/internal/config"
)

type ServiceContext struct {
	Config config.Config

	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel

	L2NftModel            nft.L2NftModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftHistoryModel     nft.L2NftHistoryModel

	TxDetailModel      tx.TxDetailModel
	TxModel            tx.TxModel
	BlockModel         block.BlockModel
	MempoolDetailModel mempool.MempoolTxDetailModel
	MempoolModel       mempool.MempoolModel
	L2AssetInfoModel   assetInfo.AssetInfoModel

	SysConfigModel sysconfig.SysconfigModel
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
		Config:                c,
		AccountModel:          account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		L2NftModel:            nft.NewL2NftModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, c.CacheRedis, gormPointer),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(conn, c.CacheRedis, gormPointer),
		TxDetailModel:         tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		TxModel:               tx.NewTxModel(conn, c.CacheRedis, gormPointer, redisConn),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		MempoolDetailModel:    mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		L2AssetInfoModel:      assetInfo.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:        sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
	}
}

/*
func (s *ServiceContext) Run() {
	mempoolTxs, err := s.MempoolModel.GetAllMempoolTxsList()
	if err != nil {
		errInfo := fmt.Sprintf("[CommitterTask] => [MempoolModel.GetAllMempoolTxsList] mempool query error:%s", err.Error())
		logx.Error(errInfo)
		return
	}
	if len(mempoolTxs) == 0 {
		logx.Info("[CommitterTask] No new mempool transactions")
		return
	} else {
		s.CommitterTask(mempoolTxs)
	}
}
func (s *ServiceContext) InitMerkleTree() (err error) {
	accounts, err := s.AccountModel.GetAllAccounts()
	if err != nil {
		return err
	}
	generalAssets, err := s.AccountAssetModel.GetAllAccountAssets()
	if err != nil {
		return err
	}
	liquidityAssets, err := s.LiquidityAssetModel.GetAllLiquidityAssets()
	if err != nil {
		return err
	}
	lockAssets, err := s.LockAssetModel.GetAllLockedAssets()
	if err != nil {
		return err
	}
	s.GlobalState, err = smt.ConstructGlobalState(accounts, generalAssets, liquidityAssets, lockAssets)
	if err != nil {
		return err
	}
	return nil
}
func (s *ServiceContext) CommitterTask(mempoolTxs []*mempool.MempoolTx) {
	//
	logx.Info("CommitterTask")
}
*/
