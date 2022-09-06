package info

import (
	"context"
	asset2 "github.com/bnb-chain/zkbnb/dao/asset"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetWithdrawGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawGasFeeLogic {
	return &GetWithdrawGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawGasFeeLogic) GetWithdrawGasFee(req *types.ReqGetWithdrawGasFee) (*types.GasFee, error) {
	resp := &types.GasFee{}

	asset, err := l.svcCtx.MemCache.GetAssetByIdWithFallback(int64(req.AssetId), func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetById(int64(req.AssetId))
	})
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}

	if asset.IsGasAsset != asset2.IsGasAsset {
		logx.Errorf("not gas asset id: %d", asset.AssetId)
		return nil, types2.AppErrInvalidGasAsset
	}
	sysGasFee, err := l.svcCtx.MemCache.GetSysConfigWithFallback(types2.SysGasFee, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(types2.SysGasFee)
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	sysGasFeeBigInt, isValid := new(big.Int).SetString(sysGasFee.Value, 10)
	if !isValid {
		logx.Errorf("parse sys gas fee err: %s", err.Error())
		return nil, types2.AppErrInternal
	}
	// if asset id == BNB, just return
	if asset.AssetId == types2.BNBAssetId {
		resp.GasFee = sysGasFeeBigInt.String()
		return resp, nil
	}
	// if not, try to compute the gas amount based on USD
	assetPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, asset.AssetSymbol)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	bnbPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, "BNB")
	if err != nil {
		return nil, types2.AppErrInternal
	}
	bnbDecimals, _ := new(big.Int).SetString(types2.BNBDecimalsStr, 10)
	assetDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(asset.Decimals)), nil)
	// bnbPrice * bnbAmount * assetDecimals / (10^18 * assetPrice)
	left := ffmath.FloatMul(ffmath.FloatMul(big.NewFloat(bnbPrice), ffmath.IntToFloat(sysGasFeeBigInt)), ffmath.IntToFloat(assetDecimals))
	right := ffmath.FloatMul(ffmath.IntToFloat(bnbDecimals), big.NewFloat(assetPrice))
	gasFee, err := common.CleanPackedFee(ffmath.FloatToInt(ffmath.FloatDiv(left, right)))
	if err != nil {
		logx.Errorf("unable to clean packed fee: %s", err.Error())
		return nil, types2.AppErrInternal
	}
	resp.GasFee = gasFee.String()
	return resp, nil
}
