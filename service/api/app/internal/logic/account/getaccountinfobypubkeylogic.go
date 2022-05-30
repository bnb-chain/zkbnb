package account

import (
	"context"
	"math/big"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByPubKeyLogic struct {
	logx.Logger
	ctx            context.Context
	svcCtx         *svc.ServiceContext
	accountHistory accounthistory.AccountHistory
	account        account.AccountModel
	globalRPC      globalrpc.GlobalRPC
	l2asset        l2asset.L2asset
}

func NewGetAccountInfoByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByPubKeyLogic {
	return &GetAccountInfoByPubKeyLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		accountHistory: accounthistory.New(svcCtx.Config),
		account:        account.New(svcCtx.Config),
		globalRPC:      globalrpc.New(svcCtx.Config),
		l2asset:        l2asset.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByPubKeyLogic) GetAccountInfoByPubKey(req *types.ReqGetAccountInfoByPubKey) (resp *types.RespGetAccountInfoByPubKey, err error) {
	resp.AccountPk = req.AccountPk
	if utils.CheckAccountPK(req.AccountPk) {
		logx.Error("[CheckAccountPK] param:%v", req.AccountPk)
		return nil, errcode.ErrInvalidParam
	}
	accountHistory, err := l.accountHistory.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Error("[GetAccountByPk] err:%v", err)
		return nil, err
	}
	resp.AccountIndex = uint32(accountHistory.AccountIndex)
	accountInfo, err := l.globalRPC.GetLatestAccountInfo(int64(accountHistory.AccountIndex))
	if err != nil {
		logx.Error("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	resp.AccountName = accountInfo.AccountName
	////
	l2AssetsList, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Error("[GetL2AssetsList] err:%v", err)
		return nil, err
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
				// Balance: accountInfo.AssetInfo[v.AssetId].Balance,
			})
	}
	resp.AssetsAccount = assets

	return resp, nil
}
