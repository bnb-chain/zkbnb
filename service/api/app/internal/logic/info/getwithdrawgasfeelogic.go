package info

import (
	"context"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"math"

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
		price:   price.New(svcCtx),
		l2asset: l2asset.New(svcCtx),
	}
}

func (l *GetWithdrawGasFeeLogic) GetWithdrawGasFee(req *types.ReqGetWithdrawGasFee) (*types.RespGetWithdrawGasFee, error) {
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(req.AssetId)
	if err != nil {
		logx.Errorf("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	withdrawL2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(req.WithdrawAssetId)
	if err != nil {
		logx.Errorf("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	price, err := l.price.GetCurrencyPrice(l.ctx, l2Asset.AssetSymbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] L2Symbol:%v, err:%v", l2Asset.AssetSymbol, err)
		return nil, err
	}
	withdrawPrice, err := l.price.GetCurrencyPrice(l.ctx, withdrawL2Asset.AssetSymbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] L2Symbol:%v, err:%v", withdrawL2Asset.AssetSymbol, err)
		return nil, err
	}
	// TODO: integer overflow
	resp := &types.RespGetWithdrawGasFee{}
	WithdrawGasFee := price * float64(req.WithdrawAmount) * math.Pow(10, -float64(l2Asset.Decimals)) * 0.001 / withdrawPrice
	minNum := math.Pow(10, -float64(l2Asset.Decimals))
	WithdrawGasFee = truncate(WithdrawGasFee, int64(l2Asset.Decimals))
	if WithdrawGasFee < minNum {
		WithdrawGasFee = minNum
	}
	WithdrawGasFee = WithdrawGasFee * math.Pow(10, float64(l2Asset.Decimals))
	resp.WithdrawGasFee = uint64(WithdrawGasFee)
	return resp, nil
}
