package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetBalanceByAssetIdAndAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetBalanceByAssetIdAndAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceByAssetIdAndAccountNameLogic {
	return &GetBalanceByAssetIdAndAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetBalanceByAssetIdAndAccountNameLogic) GetBalanceByAssetIdAndAccountName(req *types.ReqGetBlanceByAssetIdAndAccountName) (*types.RespGetBlanceInfoByAssetIdAndAccountName, error) {
	if utils.CheckAssetId(req.AssetId) {
		logx.Errorf("invalid AssetId: %s", req.AssetId)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AssetId")
	}

	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("invalid AccountName: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}
	accountName := utils.FormatAccountName(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		logx.Errorf("invalid AccountName: %s", accountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(accountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp := &types.RespGetBlanceInfoByAssetIdAndAccountName{}
	assets, err := l.globalRPC.GetLatestAssetsListByAccountIndex(l.ctx, uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("fail to get asset info for account: %d from rpc, err: %s", account.AccountIndex, err.Error())
		if err == errorcode.RpcErrNotFound {
			return resp, nil
		}
		return nil, errorcode.AppErrInternal
	}

	for _, asset := range assets {
		if req.AssetId == asset.AssetId {
			resp.Balance = asset.Balance
		}
	}
	return resp, nil
}
