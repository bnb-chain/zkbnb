package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetNextNonceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNextNonceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNextNonceLogic {
	return &GetNextNonceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNextNonceLogic) GetNextNonce(req *types.ReqGetNextNonce) (*types.NextNonce, error) {
	bc := core.NewBlockChainForDryRun(l.svcCtx.AccountModel, l.svcCtx.LiquidityModel, l.svcCtx.NftModel, l.svcCtx.MempoolModel,
		l.svcCtx.RedisCache)
	nonce, err := bc.GetPendingNonce(int64(req.AccountIndex))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	return &types.NextNonce{
		Nonce: uint64(nonce),
	}, nil
}
