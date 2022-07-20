package commglobalmap

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
)

//go:generate mockgen -source api.go -destination api_mock.go -package commglobalmap

type GlobalAssetInfo struct {
	AccountIndex   int64
	AssetId        int64
	AssetType      int64
	ChainId        int64
	BaseBalanceEnc string
}

type Model interface {
	DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error
	GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error)
	SetLatestAccountInfoInToCache(ctx context.Context, accountIndex int64) error
	GetLatestAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		mempoolModel:         mempool.NewMempoolModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		mempoolTxDetailModel: mempool.NewMempoolDetailModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		accountModel:         account.NewAccountModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		liquidityModel:       liquidity.NewLiquidityModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		redisConnection:      svcCtx.RedisConnection,
		offerModel:           nft.NewOfferModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		cache:                svcCtx.Cache,
	}
}
