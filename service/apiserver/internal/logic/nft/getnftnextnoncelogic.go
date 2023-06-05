package nft

import (
	"context"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNftNextNonceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNftNextNonceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNftNextNonceLogic {
	return &GetNftNextNonceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNftNextNonceLogic) GetNftNextNonce(req *types.ReqGetNftNextNonce) (resp *types.NextNonce, err error) {
	l2NftMetadataHistory, err := l.svcCtx.NftMetadataHistoryModel.GetL2NftMetadataHistory(req.NftIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNftNotFound
		}
		return nil, types2.AppErrInternal
	}
	return &types.NextNonce{
		Nonce: uint64(l2NftMetadataHistory.Nonce) + 1,
	}, nil
}
