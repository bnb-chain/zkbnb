package nft

import (
	"context"
	"github.com/bnb-chain/zkbnb/common"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNftByNftIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNftByNftIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNftByNftIndexLogic {
	return &GetNftByNftIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNftByNftIndexLogic) GetNftByNftIndex(req *types.ReqGetNft) (resp *types.NftEntity, err error) {
	resp = &types.NftEntity{Nft: nil}
	nft, err := l.svcCtx.NftModel.GetNft(req.NftIndex)
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
		return nil, types2.AppErrNftNotFound
	}
	creatorL1Address, _ := l.svcCtx.MemCache.GetL1AddressByIndex(nft.CreatorAccountIndex)
	ownerL1Address, _ := l.svcCtx.MemCache.GetL1AddressByIndex(nft.OwnerAccountIndex)
	resp.Nft = &types.Nft{
		Index:               nft.NftIndex,
		CreatorAccountIndex: nft.CreatorAccountIndex,
		CreatorL1Address:    creatorL1Address,
		OwnerAccountIndex:   nft.OwnerAccountIndex,
		OwnerL1Address:      ownerL1Address,
		ContentHash:         nft.NftContentHash,
		RoyaltyRate:         nft.RoyaltyRate,
		CollectionId:        nft.CollectionId,
		IpfsId:              common.GenerateCid(nft.NftContentHash),
	}
	histoty, err := l.svcCtx.NftMetadataHistoryModel.GetL2NftMetadataHistory(req.NftIndex)
	if err == nil {
		resp.Nft.IpnsId = histoty.IpnsId
		resp.Nft.Metadata = histoty.Metadata
		resp.Nft.MutableAttributes = histoty.Mutable
	}
	return resp, nil
}
