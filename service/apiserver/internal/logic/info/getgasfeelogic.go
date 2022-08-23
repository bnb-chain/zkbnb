package info

import (
	"context"
	"math/big"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/common/util"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	return &GetGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (*types.GasFee, error) {
	resp := &types.GasFee{}

	asset, err := l.svcCtx.MemCache.GetAssetByIdWithFallback(int64(req.AssetId), func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetById(int64(req.AssetId))
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	if asset.IsGasAsset != asset.IsGasAsset {
		logx.Errorf("not gas asset id: %d", asset.AssetId)
		return nil, errorcode.AppErrInvalidGasAsset
	}
	sysGasFee, err := l.svcCtx.MemCache.GetSysConfigWithFallback(sysConfigName.SysGasFee, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(sysConfigName.SysGasFee)
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	sysGasFeeBigInt, isValid := new(big.Int).SetString(sysGasFee.Value, 10)
	if !isValid {
		logx.Errorf("parse sys gas fee err: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	// if asset id == BNB, just return
	if asset.AssetId == commonConstant.BNBAssetId {
		resp.GasFee = sysGasFeeBigInt.String()
		return resp, nil
	}
	// if not, try to compute the gas amount based on USD
	assetPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	bnbPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, "BNB")
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	bnbDecimals, _ := new(big.Int).SetString(commonConstant.BNBDecimalsStr, 10)
	assetDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(asset.Decimals)), nil)
	// bnbPrice * bnbAmount * assetDecimals / (10^18 * assetPrice)
	left := ffmath.FloatMul(ffmath.FloatMul(big.NewFloat(bnbPrice), ffmath.IntToFloat(sysGasFeeBigInt)), ffmath.IntToFloat(assetDecimals))
	right := ffmath.FloatMul(ffmath.IntToFloat(bnbDecimals), big.NewFloat(assetPrice))
	gasFee, err := util.CleanPackedFee(ffmath.FloatToInt(ffmath.FloatDiv(left, right)))
	if err != nil {
		logx.Errorf("unable to clean packed fee: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	resp.GasFee = gasFee.String()
	return resp, nil
}
