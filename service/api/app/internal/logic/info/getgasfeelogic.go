package info

import (
	"context"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/price"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"math"
	"math/big"

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
	resp := &types.RespGetGasFee{}
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(l.ctx, req.AssetId)
	if err != nil {
		logx.Errorf("[GetGasFee] err:%v", err)
		return nil, err
	}
	sysGasFee, err := l.sysconf.GetSysconfigByName(l.ctx, sysconfigName.SysGasFee)
	if err != nil {
		logx.Errorf("[GetGasFee] err:%v", err)
		return nil, err
	}
	sysGasFeeBigInt, isValid := new(big.Int).SetString(sysGasFee.Value, 10)
	if !isValid {
		logx.Errorf("[GetGasFee] parse sys gas fee err:%v", err)
		return nil, err
	}
	// if asset id == BNB, just return
	if l2Asset.AssetId == commonConstant.BNBAssetId {
		resp.GasFee = sysGasFeeBigInt.String()
		return resp, nil
	}
	// if not, try to compute the gas amount based on USD
	assetPrice, err := l.price.GetCurrencyPrice(l.ctx, l2Asset.AssetSymbol)
	if err != nil {
		logx.Errorf("[GetGasFee] err:%v", err)
		return nil, err
	}
	bnbPrice, err := l.price.GetCurrencyPrice(l.ctx, "BNB")
	if err != nil {
		logx.Errorf("[GetGasFee] err:%v", err)
		return nil, err
	}
	assetInfo, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(context.Background(), req.AssetId)
	if err != nil {
		logx.Errorf("[GetGasFee] unable to get l2 asset info: %s", err.Error())
		return nil, err
	}
	bnbDecimals, _ := new(big.Int).SetString(commonConstant.BNBDecimalsStr, 10)
	assetDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(assetInfo.Decimals)), nil)
	// bnbPrice * bnbAmount * assetDecimals / (10^18 * assetPrice)
	left := ffmath.FloatMul(ffmath.FloatMul(big.NewFloat(bnbPrice), ffmath.IntToFloat(sysGasFeeBigInt)), ffmath.IntToFloat(assetDecimals))
	right := ffmath.FloatMul(ffmath.IntToFloat(bnbDecimals), big.NewFloat(assetPrice))
	gasFee, err := util.CleanPackedFee(ffmath.FloatToInt(ffmath.FloatDiv(left, right)))
	if err != nil {
		logx.Errorf("[GetGasFee] unable to clean packed fee: %s", err.Error())
		return nil, err
	}
	resp.GasFee = gasFee.String()
	return resp, nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func truncate(num float64, precision int64) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
