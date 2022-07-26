package commglobalmap

import (
	"context"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	commGlobalmapHandler "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

//go:generate mockgen -source api.go -destination api_mock.go -package commglobalmap

type GlobalAssetInfo struct {
	AccountIndex   int64
	AssetId        int64
	AssetType      int64
	ChainId        int64
	BaseBalanceEnc string
}

type Commglobalmap interface {
	GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error)
	SetLatestAccountInfoInToCache(ctx context.Context, accountIndex int64) error
	DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error

	GetLatestAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestLiquidityInfoForReadWithCache(ctx context.Context, pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
	GetLatestLiquidityInfoForRead(ctx context.Context, pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
	GetLatestOfferIdForWrite(ctx context.Context, accountIndex int64) (nftIndex int64, err error)
	GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetBasicAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error)

	GetLatestNftInfoForRead(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error)
	GetLatestNftInfoForReadWithCache(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error)
	SetLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error
	DeleteLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error

	GetLatestLiquidityInfoForWrite(ctx context.Context, pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
	SetLatestLiquidityInfoForWrite(ctx context.Context, pairIndex int64) error
	DeleteLatestLiquidityInfoForWriteInCache(ctx context.Context, pairIndex int64) error
}

func New(svcCtx *svc.ServiceContext) Commglobalmap {
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
