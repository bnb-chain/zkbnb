package info

import (
	"context"
	"errors"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/l2asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/sysconf"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	resp := &types.RespGetWithdrawGasFee{}
	l2Asset, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(l.ctx, req.AssetId)
	if err != nil {
		logx.Errorf("[GetGasFee] err: %s", err.Error())
		return nil, err
	}
	oAssetInfo, err := l.l2asset.GetSimpleL2AssetInfoByAssetId(context.Background(), req.AssetId)
	if err != nil {
		logx.Errorf("[GetGasFee] unable to get l2 asset info: %s", err.Error())
		return nil, err
	}
	if oAssetInfo.IsGasAsset != assetInfo.IsGasAsset {
		logx.Errorf("[GetGasFee] not gas asset id")
		return nil, errors.New("[GetGasFee] not gas asset id")
	}
	sysGasFee, err := l.sysconf.GetSysconfigByName(l.ctx, sysconfigName.SysGasFee)
	if err != nil {
		logx.Errorf("[GetGasFee] err: %s", err.Error())
		return nil, err
	}
	sysGasFeeBigInt, isValid := new(big.Int).SetString(sysGasFee.Value, 10)
	if !isValid {
		logx.Errorf("[GetGasFee] parse sys gas fee err: %s", err.Error())
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
		logx.Errorf("[GetGasFee] err: %s", err.Error())
		return nil, err
	}
	bnbPrice, err := l.price.GetCurrencyPrice(l.ctx, "BNB")
	if err != nil {
		logx.Errorf("[GetGasFee] err: %s", err.Error())
		return nil, err
	}
	bnbDecimals, _ := new(big.Int).SetString(commonConstant.BNBDecimalsStr, 10)
	assetDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(oAssetInfo.Decimals)), nil)
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
