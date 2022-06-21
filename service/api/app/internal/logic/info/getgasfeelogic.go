package info

import (
	"context"
	"math"
	"strconv"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/l2asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/sysconf"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	price   price.Price
	l2asset l2asset.L2asset
	sysconf sysconf.Sysconf
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	return &GetGasFeeLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		price:   price.New(svcCtx),
		l2asset: l2asset.New(svcCtx),
		sysconf: sysconf.New(svcCtx),
	}
}

// GetGasFee 需求文档
func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (*types.RespGetGasFee, error) {
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(req.AssetId)
	if err != nil {
		logx.Errorf("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	currencyPrice, err := l.price.GetCurrencyPrice(l.ctx, l2Asset.AssetSymbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] err:%v", err)
		return nil, err
	}
	sysGasFee, err := l.sysconf.GetSysconfigByName("SysGasFee")
	if err != nil {
		logx.Errorf("[GetSysconfigByName] err:%v", err)
		return nil, err
	}
	sysGasFeeInt, err := strconv.ParseFloat(sysGasFee.Value, 64)
	if err != nil {
		logx.Errorf("[strconv.ParseFloat] err:%v", err)
		return nil, err
	}

	ethPrice, err := l.price.GetCurrencyPrice(l.ctx, "ETH")
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] err:%v", err)
		return nil, err
	}
	// TODO: integer overflow
	resp := &types.RespGetGasFee{}
	resp.GasFee = ethPrice * sysGasFeeInt * math.Pow(10, -5) / currencyPrice
	minNum := math.Pow(10, -float64(l2Asset.Decimals))
	resp.GasFee = truncate(resp.GasFee, int64(l2Asset.Decimals))
	if resp.GasFee < minNum {
		resp.GasFee = minNum
	}
	return resp, nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func truncate(num float64, precision int64) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
