package info

import (
	"context"
	"math"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawGasFeeLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	price   price.Price
	l2asset l2asset.L2asset
}

func NewGetWithdrawGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawGasFeeLogic {
	return &GetWithdrawGasFeeLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		price:   price.New(svcCtx.Config),
		l2asset: l2asset.New(svcCtx.Config),
	}
}

func (l *GetWithdrawGasFeeLogic) GetWithdrawGasFee(req *types.ReqGetWithdrawGasFee) (resp *types.RespGetWithdrawGasFee, err error) {
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(uint32(req.AssetId))
	if err != nil {
		logx.Error("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	withdrawL2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(uint32(req.WithdrawAssetId))
	if err != nil {
		logx.Error("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	price, err := l.price.GetCurrencyPrice(l2Asset.AssetSymbol)
	if err != nil {
		logx.Error("[GetCurrencyPrice] L2Symbol:%v, err:%v", l2Asset.AssetSymbol, err)
		return nil, err
	}
	withdrawPrice, err := l.price.GetCurrencyPrice(withdrawL2Asset.AssetSymbol)
	if err != nil {
		logx.Error("[GetCurrencyPrice] L2Symbol:%v, err:%v", withdrawL2Asset.AssetSymbol, err)
		return nil, err
	}
	// TODO: integer overflow
	resp.WithdrawGasFee = price * float64(req.WithdrawAmount) * math.Pow(10, -float64(l2Asset.Decimals)) * 0.001 / withdrawPrice
	minNum := math.Pow(10, -float64(l2Asset.Decimals))
	resp.WithdrawGasFee = truncate(resp.WithdrawGasFee, int64(l2Asset.Decimals))
	if resp.WithdrawGasFee < minNum {
		resp.WithdrawGasFee = minNum
	}
	return resp, nil
}
