package globalrpc

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalrpc"
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
		globalRPC: globalrpc.NewGlobalRPC(zrpc.MustNewClient(svcCtx.Config.GlobalRpc)),
		cache:     svcCtx.Cache,
	}
}
