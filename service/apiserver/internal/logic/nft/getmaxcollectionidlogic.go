package nft

import (
	"context"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMaxCollectionIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMaxCollectionIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMaxCollectionIdLogic {
	return &GetMaxCollectionIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMaxCollectionIdLogic) GetMaxCollectionId(req *types.ReqGetMaxCollectionId) (resp *types.MaxCollectionId, err error) {
	account, err := l.svcCtx.StateFetcher.GetLatestAccount(req.AccountIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrAccountNotFound
		}
		return nil, types2.AppErrInternal
	}
	if account.CollectionNonce == 0 {
		return nil, types2.AppErrNotExistCollectionId
	}
	return &types.MaxCollectionId{
		CollectionId: account.CollectionNonce - 1,
	}, nil
}
