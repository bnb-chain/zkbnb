package nft

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNftsByAccountIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNftsByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNftsByAccountIndexLogic {
	return &GetNftsByAccountIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNftsByAccountIndexLogic) GetNftsByAccountIndex(req *types.ReqGetNftsByAccountIndex) (resp *types.RespGetNftsByAccountIndex, err error) {
	total, err := l.svcCtx.NftModel.GetAccountNftTotalCount(req.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetNftsByAccountIndex{
		Total: total,
		Nfts:  make([]*types.Nft, 0),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	nftList, err := l.svcCtx.NftModel.GetNftListByAccountIndex(req.AccountIndex, int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	for _, nftItem := range nftList {
		resp.Nfts = append(resp.Nfts, &types.Nft{
			NftIndex:            nftItem.NftIndex,
			CreatorAccountIndex: nftItem.CreatorAccountIndex,
			OwnerAccountIndex:   nftItem.OwnerAccountIndex,
			NftContentHash:      nftItem.NftContentHash,
			NftL1Address:        nftItem.NftL1Address,
			NftL1TokenId:        nftItem.NftL1TokenId,
			CreatorTreasuryRate: nftItem.CreatorTreasuryRate,
			CollectionId:        nftItem.CollectionId,
		})
	}
	return resp, nil
}
