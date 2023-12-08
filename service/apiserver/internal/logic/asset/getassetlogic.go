package asset

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

const (
	queryById     = "id"
	queryBySymbol = "symbol"

	// iconBaseUrl is used for showing icons for assets, asset owners should upload png files to the github repo
	//iconBaseUrl = "https://teststatic.zkbnbchain.org/shared/icon/coin/%s.svg"
	iconBaseUrl = "https://zkbnb-static-cdn.bnbchain.org/shared/icon/coin/%s.svg"
)

type GetAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetLogic {
	return &GetAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetLogic) GetAsset(req *types.ReqGetAsset) (resp *types.Asset, err error) {
	symbol := ""
	switch req.By {
	case queryById:
		id, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil || id < 0 {
			return nil, types2.AppErrInvalidAssetId
		}
		symbol, err = l.svcCtx.MemCache.GetAssetSymbolById(id)
		if err != nil {
			if err == types2.DbErrNotFound {
				return nil, types2.AppErrAssetNotFound
			}
			return nil, types2.AppErrInternal
		}
	case queryBySymbol:
		symbol = strings.ToUpper(req.Value)
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be symbol")
	}

	asset, err := l.svcCtx.MemCache.GetAssetBySymbolWithFallback(symbol, func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetBySymbol(symbol)
	})
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrAssetNotFound
		}
		return nil, types2.AppErrInternal
	}

	assetPrice, err := l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, symbol)
	if err != nil {
		return nil, types2.AppErrInternal
	}
	resp = &types.Asset{
		Id:         asset.AssetId,
		Name:       asset.AssetName,
		Decimals:   asset.Decimals,
		Symbol:     asset.AssetSymbol,
		Address:    asset.L1Address,
		Price:      strconv.FormatFloat(assetPrice, 'E', -1, 64),
		IsGasAsset: asset.IsGasAsset,
		Icon:       fmt.Sprintf(iconBaseUrl, strings.ToLower(asset.AssetSymbol)),
	}
	return resp, nil
}
