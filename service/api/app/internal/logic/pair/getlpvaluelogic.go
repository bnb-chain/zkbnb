package pair

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetLPValueLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetLPValueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLPValueLogic {
	return &GetLPValueLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetLPValueLogic) GetLPValue(req *types.ReqGetLPValue) (resp *types.RespGetLPValue, err error) {
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}
	lpValue, err := l.globalRPC.GetLpValue(l.ctx, req.PairIndex, req.LpAmount)
	if err != nil {
		logx.Errorf("fail to get lp from rpc for pair: %d, err: %s", req.PairIndex, err.Error())
		if err == errorcode.RpcErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp = &types.RespGetLPValue{
		AssetAId:     lpValue.AssetAId,
		AssetAAmount: lpValue.AssetAAmount,
		AssetBid:     lpValue.AssetBId,
		AssetBAmount: lpValue.AssetBAmount,
	}
	return resp, nil
}
