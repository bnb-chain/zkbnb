package transaction

import (
	"context"

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

func (l *GetNextNonceLogic) GetNextNonce(req *types.ReqGetNextNonce) (*types.RespGetNextNonce, error) {
	accountInfo, err := l.svcCtx.StateFetcher.GetLatestAccountInfo(l.ctx, int64(req.AccountIndex))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	return &types.RespGetNextNonce{
		Nonce: uint64(accountInfo.Nonce),
	}, nil
}
