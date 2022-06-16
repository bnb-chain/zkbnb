package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMaxOfferIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMaxOfferIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMaxOfferIdLogic {
	return &GetMaxOfferIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

//  NFT
func (l *GetMaxOfferIdLogic) GetMaxOfferId(in *globalRPCProto.ReqGetMaxOfferId) (*globalRPCProto.RespGetMaxOfferId, error) {
	// todo: add your logic here and delete this line

	return &globalRPCProto.RespGetMaxOfferId{}, nil
}
