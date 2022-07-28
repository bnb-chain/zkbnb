package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type GetMaxOfferIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetMaxOfferIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMaxOfferIdLogic {
	return &GetMaxOfferIdLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

//  NFT
func (l *GetMaxOfferIdLogic) GetMaxOfferId(in *globalRPCProto.ReqGetMaxOfferId) (*globalRPCProto.RespGetMaxOfferId, error) {
	// todo: add your logic here and delete this line
	nftIndex, err := l.commglobalmap.GetLatestOfferIdForWrite(l.ctx, int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetMaxOfferId{
		OfferId: uint64(nftIndex),
	}, nil
}
