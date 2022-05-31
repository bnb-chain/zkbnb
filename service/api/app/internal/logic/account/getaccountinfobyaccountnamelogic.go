package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	account account.HistoryAccount
	l2asset l2asset.L2asset
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		account: account.New(svcCtx.Config),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (resp *types.RespGetAccountInfoByAccountName, err error) {
	return resp, err
}
