package info

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"

	"github.com/bnb-chain/zkbnb/common"
	asset2 "github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx          context.Context
	svcCtx       *svc.ServiceContext
	gasFeeConfig map[int]int64
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	gasConfig, err := svcCtx.SysConfigModel.GetSysConfigByName(types2.SysGasFee)
	if err != nil {
		panic("fail to get gas fee config")
	}
	m := make(map[int]int64)
	err = json.Unmarshal([]byte(gasConfig.Value), &m)
	if err != nil {
		panic("invalid sys config of gas fee")
	}

	return &GetGasFeeLogic{
		Logger:       logx.WithContext(ctx),
		ctx:          ctx,
		svcCtx:       svcCtx,
		gasFeeConfig: m,
	}
}

func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (*types.GasFee, error) {
	gas, ok := l.gasFeeConfig[int(req.TxType)]
	if !ok {
		return nil, types2.AppErrInvalidTxType
	}

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

	resp := &types.GasFee{}
	sysGasFeeBigInt := big.NewInt(gas)

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
	bnbDecimals, _ := new(big.Int).SetString(types2.BNBDecimals, 10)
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
