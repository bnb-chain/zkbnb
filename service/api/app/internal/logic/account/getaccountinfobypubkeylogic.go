package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountInfoByPubKeyLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetAccountInfoByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByPubKeyLogic {
	return &GetAccountInfoByPubKeyLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetAccountInfoByPubKeyLogic) GetAccountInfoByPubKey(req *types.ReqGetAccountInfoByPubKey) (*types.RespGetAccountInfoByPubKey, error) {
	if !utils.ValidateAccountPk(req.AccountPk) {
		logx.Errorf("invalid AccountPk: %s", req.AccountPk)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountPk")
	}

	info, err := l.svcCtx.AccountModel.GetAccountByPk(req.AccountPk)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	account, err := l.commglobalmap.GetLatestAccountInfo(l.ctx, info.AccountIndex)
	if err != nil {
		logx.Errorf("fail to get account info: %d, err: %s", info.AccountIndex, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetAccountInfoByPubKey{
		AccountStatus: uint32(account.Status),
		AccountName:   account.AccountName,
		AccountIndex:  account.AccountIndex,
		Nonce:         account.Nonce,
		Assets:        make([]*types.AccountAsset, 0),
	}
	for _, asset := range account.AssetInfo {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  uint32(asset.AssetId),
			Balance:                  asset.Balance.String(),
			LpAmount:                 asset.LpAmount.String(),
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized.String(),
		})
	}
	return resp, nil
}
