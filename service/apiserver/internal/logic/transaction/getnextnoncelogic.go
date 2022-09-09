package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
	bc := core.NewBlockChainForDryRun(l.svcCtx.AccountModel, l.svcCtx.LiquidityModel, l.svcCtx.NftModel,
		l.svcCtx.MempoolModel, l.svcCtx.AssetModel, l.svcCtx.SysConfigModel, l.svcCtx.RedisCache)
	nonce, err := bc.StateDB().GetPendingNonce(int64(req.AccountIndex))
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}
	return &types.NextNonce{
		Nonce: uint64(nonce),
	}, nil
}
