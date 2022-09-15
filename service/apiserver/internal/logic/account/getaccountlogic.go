package account

import (
	"context"
	"math/big"
	"sort"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

const (
	queryByIndex = "index"
	queryByName  = "name"
	queryByPk    = "pk"
)

type GetAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountLogic {
	return &GetAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountLogic) GetAccount(req *types.ReqGetAccount) (resp *types.Account, err error) {
	index := int64(0)
	switch req.By {
	case queryByIndex:
		index, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil || index < 0 {
			return nil, types2.AppErrInvalidParam.RefineError("invalid value for account index")
		}
	case queryByName:
		index, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case queryByPk:
		index, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be index|name|pk")
	}

	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}

	account, err := l.svcCtx.StateFetcher.GetLatestAccount(index)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNotFound
		}
		return nil, types2.AppErrInternal
	}

	maxAssetId, err := l.svcCtx.AssetModel.GetMaxAssetId()
	if err != nil {
		return nil, types2.AppErrInternal
	}

	resp = &types.Account{
		Index:  account.AccountIndex,
		Status: uint32(account.Status),
		Name:   account.AccountName,
		Pk:     account.PublicKey,
		Nonce:  account.Nonce,
		Assets: make([]*types.AccountAsset, 0),
		Lps:    make([]*types.AccountLp, 0),
	}
	for _, asset := range account.AssetInfo {
		if asset.AssetId > maxAssetId {
			continue //it is used for offer related, or empty balance; max ip id should be less than max asset id
		}
		if (asset.Balance == nil || asset.Balance.Cmp(big.NewInt(0)) == 0) &&
			(asset.LpAmount == nil || asset.LpAmount.Cmp(big.NewInt(0)) == 0) {
			continue
		}
		if asset.Balance != nil && asset.Balance.Cmp(big.NewInt(0)) > 0 {
			var assetName, assetSymbol string
			var assetPrice float64
			assetName, err = l.svcCtx.MemCache.GetAssetNameById(asset.AssetId)
			if err != nil {
				return nil, types2.AppErrInternal
			}
			assetSymbol, err = l.svcCtx.MemCache.GetAssetSymbolById(asset.AssetId)
			if err != nil {
				return nil, types2.AppErrInternal
			}
			assetPrice, err = l.svcCtx.PriceFetcher.GetCurrencyPrice(l.ctx, assetSymbol)
			if err != nil {
				return nil, types2.AppErrInternal
			}
			resp.Assets = append(resp.Assets, &types.AccountAsset{
				Id:      uint32(asset.AssetId),
				Name:    assetName,
				Balance: asset.Balance.String(),
				Price:   strconv.FormatFloat(assetPrice, 'E', -1, 64),
			})
		}
		if asset.LpAmount != nil && asset.LpAmount.Cmp(big.NewInt(0)) > 0 {
			resp.Lps = append(resp.Lps, &types.AccountLp{
				Index:  uint32(asset.AssetId),
				Amount: asset.LpAmount.String(),
			})
		}
	}

	sort.Slice(resp.Assets, func(i, j int) bool {
		return resp.Assets[i].Id < resp.Assets[j].Id
	})

	return resp, nil
}
