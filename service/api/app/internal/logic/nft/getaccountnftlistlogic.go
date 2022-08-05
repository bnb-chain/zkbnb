package nft

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountNftListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountNftListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountNftListLogic {
	return &GetAccountNftListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountNftListLogic) GetAccountNftList(req *types.ReqGetAccountNftList) (*types.RespGetAccountNftList, error) {
	total, err := l.svcCtx.NftModel.GetAccountNftTotalCount(req.AccountIndex)
	if err != nil {
		logx.Errorf("[GetAccountNftList] get account nft total count error: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp := &types.RespGetAccountNftList{
		Total: total,
		Nfts:  make([]*types.Nft, 0),
	}
	if total == 0 || total < int64(req.Offset) {
		return resp, nil
	}

	nftList, err := l.svcCtx.NftModel.GetNftListByAccountIndex(req.AccountIndex, int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetAccountNftList] get nft list by account error: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
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
