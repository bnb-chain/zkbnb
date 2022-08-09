package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetNextNonceLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetNextNonceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNextNonceLogic {
	return &GetNextNonceLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetNextNonceLogic) GetNextNonce(req *types.ReqGetNextNonce) (*types.RespGetNextNonce, error) {
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(l.ctx, int64(req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	return &types.RespGetNextNonce{
		Nonce: uint64(accountInfo.Nonce),
	}, nil
}
