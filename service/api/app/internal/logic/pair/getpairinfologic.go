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

type GetPairInfoLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairInfoLogic {
	return &GetPairInfoLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetPairInfoLogic) GetPairInfo(req *types.ReqGetPairInfo) (*types.RespGetPairInfo, error) {
	if !utils.ValidatePairIndex(req.PairIndex) {
		logx.Errorf("invalid PairIndex: %d", req.PairIndex)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid PairIndex")
	}

	pair, err := l.globalRPC.GetPairInfo(l.ctx, req.PairIndex)
	if err != nil {
		logx.Errorf("fail to get pair info from rpc for pair: %d, err: %s", req.PairIndex, err.Error())
		if err == errorcode.RpcErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetPairInfo{
		AssetAId:      pair.AssetAId,
		AssetAAmount:  pair.AssetAAmount,
		AssetBId:      pair.AssetBId,
		AssetBAmount:  pair.AssetBAmount,
		TotalLpAmount: pair.LpAmount,
	}
	return resp, nil
}
