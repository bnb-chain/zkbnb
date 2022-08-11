package info

import (
	"context"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	price  price.Price
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	return &GetGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		price:  price.New(svcCtx),
	}
}

func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (*types.RespGetGasFee, error) {
	resp := &types.RespGetGasFee{}

	assetInfo, err := l.svcCtx.L2AssetModel.GetAssetByAssetId(int64(req.AssetId))
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	if assetInfo.IsGasAsset != assetInfo.IsGasAsset {
		logx.Errorf("not gas asset id: %d", assetInfo.AssetId)
		return nil, errorcode.AppErrInvalidGasAsset
	}
	sysGasFee, err := l.svcCtx.SysConfigModel.GetSysConfigByName(sysConfigName.SysGasFee)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	sysGasFeeBigInt, isValid := new(big.Int).SetString(sysGasFee.Value, 10)
	if !isValid {
		logx.Errorf("parse sys gas fee err: %s", err.Error())
		return nil, errorcode.AppErrInternal
	}
	// if asset id == BNB, just return
	if assetInfo.AssetId == commonConstant.BNBAssetId {
		resp.GasFee = sysGasFeeBigInt.String()
		return resp, nil
	}
	// if not, try to compute the gas amount based on USD
	assetPrice, err := l.price.GetCurrencyPrice(l.ctx, assetInfo.AssetSymbol)
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	bnbPrice, err := l.price.GetCurrencyPrice(l.ctx, "BNB")
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	bnbDecimals, _ := new(big.Int).SetString(commonConstant.BNBDecimalsStr, 10)
	assetDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(assetInfo.Decimals)), nil)
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
