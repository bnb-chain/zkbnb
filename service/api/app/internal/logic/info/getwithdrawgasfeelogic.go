package info

import (
	"context"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"math"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawGasFeeLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	price   price.Price
	l2asset l2asset.L2asset
	sysconf sysconf.Sysconf
}

func NewGetWithdrawGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawGasFeeLogic {
	return &GetWithdrawGasFeeLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		price:   price.New(svcCtx),
		l2asset: l2asset.New(svcCtx),
		sysconf: sysconf.New(svcCtx),
	}
}

//todo modify 【now function copy from service/api/app/internal/logic/info/getgasfeelogic.go:38】
func (l *GetWithdrawGasFeeLogic) GetWithdrawGasFee(req *types.ReqGetWithdrawGasFee) (*types.RespGetWithdrawGasFee, error) {
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(req.AssetId)
	if err != nil {
		logx.Errorf("[GetSimpleL2AssetInfoByAssetId] err:%v", err)
		return nil, err
	}
	SymbolPrice, err := l.price.GetCurrencyPrice(l.ctx, l2Asset.AssetSymbol)
	if err != nil {
		logx.Errorf("[GetCurrencyPrice] L2Symbol:%v, err:%v", l2Asset.AssetSymbol, err)
		return nil, err
	}

	// TODO: integer overflow
	ethPrice, err := l.price.GetCurrencyPrice(l.ctx, "ETH")
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
	resp := &types.RespGetWithdrawGasFee{}
	WithdrawGasFee := ethPrice * sysGasFeeInt * math.Pow(10, -5) / SymbolPrice
	minNum := math.Pow(10, -float64(l2Asset.Decimals))
	WithdrawGasFee = truncate(WithdrawGasFee, int64(l2Asset.Decimals))
	if WithdrawGasFee < minNum {
		WithdrawGasFee = minNum
	}
	WithdrawGasFee = WithdrawGasFee * math.Pow(10, float64(l2Asset.Decimals))
	resp.WithdrawGasFee = uint64(WithdrawGasFee)
	return resp, nil
}
