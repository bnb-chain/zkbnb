package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAssetIdAndPubKeyLogic struct {
	logx.Logger
	ctx            context.Context
	svcCtx         *svc.ServiceContext
	accountHistory accounthistory.AccountHistory
	account        account.AccountModel
	globalRPC      globalrpc.GlobalRPC
}

func NewGetAccountInfoByAssetIdAndPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAssetIdAndPubKeyLogic {
	return &GetAccountInfoByAssetIdAndPubKeyLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		accountHistory: accounthistory.New(svcCtx.Config),
		account:        account.New(svcCtx.Config),
		globalRPC:      globalrpc.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByAssetIdAndPubKeyLogic) GetAccountInfoByAssetIdAndPubKey(req *types.ReqGetAccountInfoByAssetIdAndPubKey) (resp *types.RespGetAccountInfoByAssetIdAndPubKey, err error) {
	if utils.CheckAccountPK(req.AccountPk) {
		logx.Error("[CheckAccountPK] param:%v", req.AccountPk)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAssetId(req.AssetId) {
		logx.Error("[CheckAssetId] param:%v", req.AssetId)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Error("[GetAccountByPk] err:%v", err)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckAccountIndex(uint32(account.AccountIndex)) {
		logx.Error("[CheckAccountIndex] param:%v", account.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.globalRPC.GetLatestAccountInfo(int64(account.AccountIndex))
	if err != nil {
		logx.Error("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	resp.AssetId = req.AssetId
	resp.AccountIndex = uint32(accountInfo.AccountIndex)
	resp.AccountName = accountInfo.AccountName
	resp.AccountPk = accountInfo.PublicKey
	// resp.BalanceEnc = accountInfo.AssetInfo[int64(req.AssetId)].Balance
	return resp, nil
}
