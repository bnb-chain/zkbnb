package commglobalmap

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
)

type GlobalAssetInfo struct {
	AccountIndex   int64
	AssetId        int64
	AssetType      int64
	ChainId        int64
	BaseBalanceEnc string
}

type Commglobalmap interface {
	GetLatestAccountInfo(accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error)
	GetLatestLiquidityInfoForRead(pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
	GetLatestOfferIdForWrite(accountIndex int64) (nftIndex int64, err error)
}

func New(svcCtx *svc.ServiceContext) Commglobalmap {
	return &commglobalmap{
		mempoolTxDetailModel: mempool.NewMempoolDetailModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		mempoolModel:         mempool.NewMempoolModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		AccountModel:         account.NewAccountModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		liquidityModel:       liquidity.NewLiquidityModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		redisConnection:      svcCtx.RedisConnection,
		offerModel:           nft.NewOfferModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
	}
}
