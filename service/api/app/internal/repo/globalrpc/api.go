package globalrpc

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
	"github.com/zeromicro/go-zero/zrpc"
)

type GlobalRPC interface {
	SendTx(ctx context.Context, txType uint32, txInfo string) (string, error)
	GetLpValue(ctx context.Context, pairIndex uint32, lpAmount string) (*globalRPCProto.RespGetLpValue, error)
	GetPairInfo(ctx context.Context, pairIndex uint32) (*globalRPCProto.RespGetLatestPairInfo, error)
	GetSwapAmount(ctx context.Context, pairIndex, assetId uint64, assetAmount string, isFrom bool) (string, uint32, error)
	GetNextNonce(ctx context.Context, accountIndex uint32) (uint64, error)
	GetLatestAssetsListByAccountIndex(ctx context.Context, accountIndex uint32) ([]*globalrpc.AssetResult, error)
	GetLatestAccountInfoByAccountIndex(ctx context.Context, accountIndex int64) (*globalrpc.RespGetLatestAccountInfoByAccountIndex, error)
	GetMaxOfferId(ctx context.Context, accountIndex uint32) (uint64, error)
	SendMintNftTx(ctx context.Context, txInfo string) (int64, error)
	SendCreateCollectionTx(ctx context.Context, txInfo string) (int64, error)

	SendAddLiquidityTx(ctx context.Context, txInfo string) (string, error)
	SendAtomicMatchTx(ctx context.Context, txInfo string) (string, error)
	SendCancelOfferTx(ctx context.Context, txInfo string) (string, error)
	SendRemoveLiquidityTx(ctx context.Context, txInfo string) (string, error)
	SendSwapTx(ctx context.Context, txInfo string) (string, error)
	SendTransferNftTx(ctx context.Context, txInfo string) (string, error)
	SendTransferTx(ctx context.Context, txInfo string) (string, error)
	SendWithdrawNftTx(ctx context.Context, txInfo string) (string, error)
	SendWithdrawTx(ctx context.Context, txInfo string) (string, error)
}

func New(svcCtx *svc.ServiceContext, ctx context.Context) GlobalRPC {
	return &globalRPC{
		AccountModel:        account.NewAccountModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		AccountHistoryModel: account.NewAccountHistoryModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		MempoolModel:        mempool.NewMempoolModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		MempoolDetailModel:  mempool.NewMempoolDetailModel(svcCtx.Conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
		RedisConnection:     svcCtx.RedisConn,
		globalRPC:           globalrpc.NewGlobalRPC(zrpc.MustNewClient(svcCtx.Config.GlobalRpc)),
		cache:               svcCtx.Cache,
	}
}
