package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetBalanceByAssetIdAndAccountNameLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetBalanceByAssetIdAndAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceByAssetIdAndAccountNameLogic {
	return &GetBalanceByAssetIdAndAccountNameLogic{
		Logger:        logx.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetBalanceByAssetIdAndAccountNameLogic) GetBalanceByAssetIdAndAccountName(req *types.ReqGetBlanceByAssetIdAndAccountName) (*types.RespGetBlanceInfoByAssetIdAndAccountName, error) {
	if !utils.ValidateAssetId(req.AssetId) {
		logx.Errorf("invalid AssetId: %s", req.AssetId)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetId")
	}

	if !utils.ValidateAccountName(req.AccountName) {
		logx.Errorf("invalid AccountName: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp := &types.RespGetBlanceInfoByAssetIdAndAccountName{}
	accountInfo, err := l.commglobalmap.GetLatestAccountInfoWithCache(l.ctx, int64(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}

	for _, asset := range accountInfo.AssetInfo {
		if req.AssetId == uint32(asset.AssetId) {
			resp.Balance = asset.Balance.String()
		}
	}
	return resp, nil
}
