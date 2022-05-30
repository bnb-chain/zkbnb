package account

import (
	"context"
	"math/big"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx            context.Context
	svcCtx         *svc.ServiceContext
	accountHistory accounthistory.AccountHistory
	l2asset        l2asset.L2asset
	globalRPC      globalrpc.GlobalRPC
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		accountHistory: accounthistory.New(svcCtx.Config),
		l2asset:        l2asset.New(svcCtx.Config),
		globalRPC:      globalrpc.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (resp *types.RespGetAccountInfoByAccountName, err error) {
	resp.AssetsAccount = make([]*types.Asset, 0)
	if utils.CheckAccountName(req.AccountName) {
		logx.Error("[CheckAccountName] req.AccountName:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountName := utils.FormatSting(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		logx.Error("[CheckFormatAccountName] accountName:%v", accountName)
		return nil, errcode.ErrInvalidParam
	}
	resp.AccountName = accountName
	account, err := l.accountHistory.GetAccountByAccountName(accountName)
	if err != nil {
		logx.Error("[GetAccountByAccountName] accountName:%v, err:%v", accountName, err)
		return nil, err
	}
	resp.AccountIndex = uint32(account.AccountIndex)
	resp.AccountPk, resp.AssetsAccount, err = l.getLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Error("[getLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	return resp, nil
}

func (l *GetAccountInfoByAccountNameLogic) getLatestAccountInfoByAccountIndex(accountIndex uint32) (
	string, []*types.Asset, error) {
	if utils.CheckAccountIndex(accountIndex) {
		logx.Error("[CheckAccountIndex] param:%v", accountIndex)
		return "", nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.globalRPC.GetLatestAccountInfo(int64(accountIndex))
	if err != nil {
		logx.Error("[CheckAccountName] err:%v", err)
		return "", nil, err
	}
	l2AssetsList, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Error("[GetL2AssetsList] err:%v", err)
		return "", nil, err
	}
	var assets []*types.Asset
	for _, v := range l2AssetsList {
		if accountInfo.AssetInfo[v.AssetId] == nil {
			accountInfo.AssetInfo[v.AssetId] = &commonAsset.AccountAsset{
				Balance: big.NewInt(0),
			}
		}
		assets = append(assets,
			&types.Asset{
				// TODO: int64 to uint16 is dangerous
				AssetId: uint16(v.AssetId),
				Balance: accountInfo.AssetInfo[v.AssetId].Balance.String(),
			})
	}
	return accountInfo.PublicKey, assets, nil
}
