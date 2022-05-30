package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetL1AmountListLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetL1AmountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetL1AmountListLogic {
	return &GetL1AmountListLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config),
	}
}

func (l *GetL1AmountListLogic) GetL1AmountList(req *types.ReqGetL1AmountList) (resp *types.RespGetL1AmountList, err error) {
	// TODO: globalRPC.GetLatestL1AmountList
	amounts, err := l.globalRPC.GetLatestL1AmountList()
	if err != nil {
		logx.Error("[GetLatestL1AmountList] err:%v", err)
		return nil, err
	}
	resp.Amounts = make([]*types.AmountInfo, 0)
	for _, l1Asset := range amounts {
		resp.Amounts = append(resp.Amounts, &types.AmountInfo{
			AssetId:     int(l1Asset.AssetId),
			TotalAmount: l1Asset.TotalAmount,
		})
	}
	return resp, nil
}
