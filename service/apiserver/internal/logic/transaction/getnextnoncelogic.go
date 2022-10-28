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
	bc, err := core.NewBlockChainForDryRun(l.svcCtx.AccountModel, l.svcCtx.NftModel,
		l.svcCtx.TxPoolModel, l.svcCtx.AssetModel, l.svcCtx.SysConfigModel, l.svcCtx.RedisCache)
	if err != nil {
		return nil, err
	}
	nonce, err := bc.StateDB().GetPendingNonce(int64(req.AccountIndex))
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrAccountNonceNotFound
		}
		return nil, types2.AppErrInternal
	}
	return &types.NextNonce{
		Nonce: uint64(nonce),
	}, nil
}
