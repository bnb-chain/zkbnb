package nft

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetMaxOfferIdLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetMaxOfferIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMaxOfferIdLogic {
	return &GetMaxOfferIdLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetMaxOfferIdLogic) GetMaxOfferId(req *types.ReqGetMaxOfferId) (resp *types.RespGetMaxOfferId, err error) {
	offerId, err := l.globalRPC.GetMaxOfferId(l.ctx, uint32(req.AccountIndex))
	if err != nil {
		logx.Errorf("[GetMaxOfferId] err:%v", err)
		return nil, err
	}
	return &types.RespGetMaxOfferId{OfferId: offerId}, nil
}
